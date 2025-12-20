---
name: locker-package-design
overview: 在 go-halo 项目中新增一个基于 string key 的可扩展 locker 接口及内存实现，支持 TryLock、带超时的 Lock 和 Release。
todos:
  - id: define-interface
    content: 在 locker 包中定义 Locker 接口和构造函数 `NewMemoryLocker`。
    status: pending
  - id: implement-memory-locker
    content: 实现基于 map+channel 的内存版 locker，包括 TryLock、Lock 和 Release。
    status: pending
    dependencies:
      - define-interface
  - id: add-tests
    content: 为内存 locker 编写并发与超时相关的单元测试。
    status: pending
    dependencies:
      - implement-memory-locker
---

# locker 包设计与实现方案

### 目标

- **新增 `locker` 包**：提供基于 `string` key 的分布式式接口设计（当前实现为进程内内存版），支持：
- **`TryLock`**：尝试获取锁，获取不到立即返回，不阻塞。
- **`Lock`**：`Lock(ctx context.Context, key string, timeout time.Duration)`，在指定超时时间内尝试获取锁，可被 `context` 取消。
- **`Unlock`**：释放指定 key 的锁。
- **接口化设计**：将 `locker` 抽象成接口，方便后续扩展其他实现（如基于 Redis、数据库等）。

### API 设计

在新文件 [`locker/locker.go`](/Users/fengjianxin/workspaces/my-opensource-project/go-halo/locker/locker.go) 中：

- **接口定义**：
- `type Locker interface {`
  - `TryLock(ctx context.Context, key string) (bool, error)`
  - `Lock(ctx context.Context, key string, timeout time.Duration) error`
  - `Unlock(key string)`
- 预留将来扩展（例如增加 `Close()`、统计方法等）。
- **构造函数**：
- `func NewMemoryLocker() Locker`：返回默认的内存实现。

### 内存实现设计

- **文件**：[`locker/memory_locker.go`](/Users/fengjianxin/workspaces/my-opensource-project/go-halo/locker/memory_locker.go)。
- **核心结构**：
- `type memoryLocker struct {`
  - `mu sync.Mutex`：保护内部 map。
  - `locks map[string]*lockEntry`
- `type lockEntry struct { ch chan struct{} }`：每个 key 一把锁，对应一个带缓冲的信号量 channel。
- **加锁实现思路**：
- 为每个 key 维护一个 **容量为 1 的 channel**，channel 中的“令牌”表示锁的可用性：
  - 创建 `lockEntry` 时先向 `ch` 发送一个空结构体，表示“锁当前空闲”。
  - **获取锁** = 从 `ch` 中接收一个值；**释放锁** = 再往 `ch` 中发送一个值。
- **`TryLock`**：
  - 通过 map 找到对应 `lockEntry`（如不存在则创建并初始化）。
  - 使用 `select` + `default` 尝试从 `ch` 非阻塞接收：
  - 收到则表示获取锁成功，返回 `true`。
  - 没有数据可读则立即返回 `false`。
  - 同时在 `select` 中监听 `ctx.Done()`，避免在创建阶段异常阻塞。
- **`Lock`**：
  - 找到或创建 `lockEntry` 后，通过 `select`：
  - `case <-ch`：成功获取锁返回。
  - `case <-time.After(timeout)`：到达超时时间返回超时错误。
  - `case <-ctx.Done()`：返回 context 错误（如被取消或 deadline 超时）。
- **`Release`**：
  - 查找对应的 `lockEntry`，如果存在则尝试往 `ch` 中发送一个值以“归还令牌”。
  - 为避免 panic，需要保证不会重复释放（多次 `Release` 不会导致写满 channel）：
  - 可以选择在调试版做严格检查，或在文档中约定由调用方保证成对调用。
  - 可选：在锁长时间未使用时清理 `locks` 中的条目（本版可先不做，保持简单）。

### 错误与返回约定

- **`TryLock`**：
- 获取锁成功：`(true, nil)`。
- 锁被占用：`(false, nil)`。
- `ctx` 被取消：返回 `false` 和 `ctx.Err()`。
- **`Lock`**：
- 成功：`nil`。
- 超时：返回自定义错误（例如 `errs.New("locker: lock timeout")`），复用项目内 `errs` 包。
- `ctx` 取消或 deadline：返回 `ctx.Err()`。

### 测试与示例

- 在 [`locker/locker_test.go`](/Users/fengjianxin/workspaces/my-opensource-project/go-halo/locker/locker_test.go) 中编写单元测试：
- **基础功能**：
  - 同一个 key 上顺序调用 `TryLock`/`Release`，验证互斥性。
  - 不同 key 之间互不影响。
- **超时与取消**：
  - 先在一个 goroutine 中持有锁，再在另一个 goroutine 中调用 `Lock`，验证：
  - 超时时返回预期错误。
  - `ctx` 取消时立刻返回。
- **并发场景**：
  - 启动多个 goroutine 争抢同一 key 的锁，确保同一时刻只有一个持有锁（可通过计数器断言）。

### 后续扩展预留

- 后续若需要 Redis / 分布式实现，可在同包下新增：
- `redis_locker.go`：实现同样的 `Locker` 接口，内部使用 Redis `SET NX PX` 等语义。
- 再通过工厂方法（例如 `NewRedisLocker(...)`) 暴露给调用方。