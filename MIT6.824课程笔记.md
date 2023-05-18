# MIT6.824课程笔记

## introduction

分布式：通过网络来协调，协同完成一致任务的一些计算机。

分布式系统
一组不共享内存和时钟处理器的集合。每个处理器都有自己的内存，之间通过各种通信网络加以实现。处理器大小和和功能不同，其中，某个站点表示特定的主机，服务器，拥有客户机所需的资源。

### 建立分布式的原因

1. 加快计算速度：将特定的计算分化成可以并发运行的子运算，分布式系统允许将这些子运算分不到不同的站点，就可以并发的进行。此外啊，还可以进行负载分配。（大量的计算机意味着）大量的并行运算，大量CPU、大量内存、以及大量磁盘在并行的运行
2. 可靠性（提供容错）：如果系统具有足够的冗余（包括硬件和数据），即使某些站点出现故障，系统也会继续运行下去。1. 分布式系统必须保证能检测到站点故障，不再使用该站点的服务，并采取适当措施修复；2. 如果有其他站点可以接管，系统必须保证操作转移的正确；3. 故障的站点恢复后，系统保证有一个顺利结合到系统的机制；
3. 一些问题天然在空间上是分布的。例如银行转账，我们假设银行A在纽约有一台服务器，银行B在伦敦有一台服务器，这就需要一种两者之间协调的方法。所以，有一些天然的原因导致系统是物理分布的。
4. 安全，通过隔离提高安全性。比如一些代码不被信任但需要交互，可以将代码分散多出运行，系统分成多个计算机，这样可以限制出错域。

### 分布式难点：

1. 并发编程
2. 复杂交互，时间依赖（同步、异步）
3. 局部错误很难调试
4. 很难获得高性能

### 基础架构： 存储 通信 计算

存储是最重要，构建一种多副本、容错的、高性能分布式存储实现；

存储和计算，目的是实现简单剪口，像非分布式存储和计算系统一样被第三方应用，但是实际上又是一个有极高的性能和容错的分布式系统。

**构建用到的工具：**

RPC（Remote Procedure Call）。RPC的目标就是掩盖我们正在不可靠网络上通信的事实。一种分布式系统中的通信方式，可以使得不同的计算机程序之间可以像本地程序一样进行相互调用，从而方便构建分布式系统。通过RPC技术，可以帮助实现以下功能 ：

1. 分布式服务调用：各个计算机之间可以调用对方提供的服务，从而实现分布式系统中的协同工作。
2. 负载均衡：可以将请求分配到多个服务器上，从而实现负载均衡，提高系统的可靠性和可用性。
3. 高并发处理：，可以实现多个请求的并发处理，提高系统的响应速度。
4. 数据共享：可以将数据共享到多个计算机之间，从而实现数据共享和协同处理。

多线程编程技术

考虑控制并发、锁

### 拓展性Sacalability

如果我用一台计算机解决了一些问题，当我买了第二台计算机，我只需要一半的时间就可以解决这些问题，或者说每分钟可以解决两倍数量的问题。两台计算机构成的系统如果有两倍性能或者吞吐，就是我说的可扩展性。

通过增加机器的方式来实现扩展，但是现实中这很难实现，需要一些架构设计来将这个可扩展性无限推进下去。

比如说并发量上去了，那需要增加服务器数量，这导致数据库成为瓶颈，又需要搭建分布式存储：
		也就是web服务器和数据库的成为瓶颈时，都能突破，拓展下去。

### 可用 | 容错性Avaliability

可用性（Availability）: 这样在特定的错误类型下，系统仍然能够正常运行，仍然可以像没有出现错误一样，为你提供完整的服务。

大型分布式系统中有一个大问题，那就是一些很罕见的问题会被放大,各个地方总是有一些小问题出现。所以大规模系统会将一些几乎不可能并且你不需要考虑的问题，变成一个持续不断的问题。

自我可恢复性（recoverability）: 如果出现了问题，服务会停止工作(通常需要做一些操作，例如将最新的数据存放在磁盘中),不再响应请求，之后有人来修复，并且在修复之后系统仍然可以正常运行，就像没有出现过问题一样

**可用和自我恢复用到的工具**

非易失存储：存放一些checkpoint或者系统状态的log在这些存储中，这样当备用电源恢复或者某人修好了电力供给，我们还是可以从硬盘中读出系统最新的状态，并从那个状态继续运行。但是非易失存储的读写的代价很高，避免频繁读写

复制：管理多副本实现容错系统

### 一致性Consistency

定义操作行为的概念。因为分布式系统考虑性能和容错，通常会使用多个副本有不同版本的 key-value 。比如在put操作中，有两个服务器，需要发给两个副本以达到同步，但是发送过程中可能客户端故障了，只有一个收到了，数据不一致了。

弱一致是指，不保证get请求可以得到最近一次完成的put请求写入的值。尽管有很多细节的工作要处理，强一致可以保证get得到的是put写入的最新的数据。所以在一个弱一致系统中，某人通过put请求写入了一个数据，但是你通过get看到的可能仍然是一个旧数据，而这个旧数据可能是很久之前写入的。  但是强一致的开销可能非常大，常常构建弱一致系统，只从最近的数据副本更新和获取数据。

### 提高性能 performance

- throughput 吞吐量：目标是吞吐量与部署的机器数成正比
- latency 低延迟：其中一台机器执行慢就会导致整个请求响应慢，这称为**尾部延迟(tail latency)**

### MapReduce基本工作方式

MapReduce的思想是，应用程序设计人员和分布式运算的使用者，只需要写简单的Map函数和Reduce函数，而不需要知道任何有关分布式的事情，MapReduce框架会处理剩下的事情。

 map-reduce经典举例即统计字母出现的次数，多个进程各自通过map函数统计获取到的数据片段的字母的出现次数；后续再通过reduce函数，汇总聚合map阶段下每个进程对各自负责的数据片段统计的字母出现次数。一旦执行了shuffle，多个reduce函数可以各自只聚合一种字母的出现总次数，彼此之间不干扰。

开销昂贵的部分即shuffle，map的结果经过shuffle按照一定的顺序整理/排序，然后才分发给不同的reduce处理。这里shuffle的操作理论比map、reduce昂贵。

## MapReduce

输入一组kv，生成一组kv。

MAP函数获取一个输入，并且生成一组中间的KV，

MapReduce库把与同一中间键i相关联的所有中间值组合在一起，通过迭代器传递给Reduce函数。

Reduce函数接受中间键i和键对应的一组值，把值合在一起，形成一个看更小的值集。一般每次只生成0或1个输出值。

<img src="D:\MyTxt\typoraPhoto\image-20230329113648794.png" alt="image-20230329113648794" style="zoom:80%;" />

### 步骤：

1. 输入文件，MapReduce库分割成M个片段，每个片段16-64MB，然后再一群机器上启动程序的副本
2. 其中一个副本是Master，其余是workers，总共有M个map任务和R个reduce任务需要分配。由主进程挑选空闲的workers分配一项任务。
3. 分配到map任务的worker读取内容（被拆分好的一部分），从输出数据中解析出KV，传递给Map函数，然后把生成的中间KV存在内存
4. 定期写入本地磁盘，并分区R个区，然后传地址给master。
5. master通知reduce任务的workers这些位置，然后用RPC读取map的workers这些数据（直接从map传到reduce workers？）。当reduceWorker读完所有中间数据后，按照键排序，以便将所有相同键的项组合。（排序是必须的，如果中间数据量大到内存不够了，用外部排序）
6. reduceWorker遍历已排序的中间数据，按照键值和对应的一组中间值传递给Reduce函数。reduce函数的结果加到此分区的输出文件中
7. 所有map和reduce任务完成后，master唤醒user程序。这时user程序中的MapReduce来return结果，一般所有分区输出文件同时输出传递给另一个MapReduce函数，不需要合并为一个文件。

### 数据结构：

每个map和reduce任务，保留存储状态（空闲，进行，完成）和工作计算机的标识（非空闲）

master是map任务中间文件位置的接受管道，master存储R个中间文件的位置和大小，map任务完成后，master接受信息更新，并且递增地传给正在进行的reduceWorkers

### 容错性：

#### workers failure

master定期ping每个worker如果有worker没有响应 或者 工作线程故障 。就标记为失败，任务重置为初始空闲状态，可以放在其他works上调度。

1. 对于map任务故障，直接全部重新执行，因为输出存储在故障计算机上，无法访问。

2. 对于reduce任务，只要接下来换机器执行，因为输出存储在全局文件系统中。

A执行map任务的时候坏了转给了B，那么对应执行reduce的workers需要重新执行，从B读取数据

对于故障 需要有很大的弹性。

#### master failure

写入master数据结构的检查点，master终止之后，可以从上一个检查点状态重新启动新的副本。

中止计算，客户端可以检查这种情况，重试

#### 存在故障的语言（没看懂）

map和reduce函数确定后，分布式产生的输出与无错误的顺序执行产生的输出相同。

主要是通过map和reduce输出的原子提交实现这一属性。reduce任务输出文件生成一个，map任务生成R个。map任务完成后，workers向master发送一条消息，包含R个文件名。reduce任务完成后，自动文件重命名为最终输出文件（如果多个worker执行相同的reduce任务，则相同的最终文件多个重命名调用）。

如果map和reduce运算是确定的，等同顺序执行，比较容易逻辑推理；但是如果存在非确定的运算，在存在非确定性运算符的情况下，特定归约任务R1的输出等同于由非确定性程序的顺序执行产生的R1的输出。然而，不同归约任务R2的输出可以对应于由非确定性程序的不同顺序执行产生的R2的输出。

举例子：M的一次输出可能被R1读取并正在执行，M的另一次输出可能被R2读取了

#### 本地化

如果失败将在输入数据的副本附近调度map任务，大多数是本地读取的，不会消耗带宽

#### 任务粒度

任务数量M和R远大于计算机数量（当然有个上限），调整动态负载平衡。 R受到用户限制，一般是机器数量的小倍数

主机必须做出O(M+R)调度决策并在内存中保持O(M∗R)状态？？？

#### 任务备份

问题：有些机器运行落后，最后几个任务要一直等着。

解决方案： 当一个MapReduce系统快完成的时候，主线程调度正在进行的任务进行备份执行，只要主执行或者备份执行有完成的，那这个任务就被标记完成。这样加了少量计算负担，能提升时间较多

### 优化拓展

#### 分区函数

用户指定M和R，使用分区函数对数据进行分区。要求平衡的分区，有时候希望相同的输出文件结束。。。

#### 序列化

保证给定分区中，中间KV按照K升序处理，方便输出文件被查找。

#### 组合函数

中间键有时候有大量重复计算，原本都需要单独发给reduce任务计算，然后合并。

现在可以在map任务的计算器上指定一个合并函数，执行合并后写入一个中间文件，发给reduce任务。这样可以提高操作速度。

这种部分合并的combiner函数和reduce函数代码一致（除了输出的位置）。

#### 输入输出类型

支持不同格式读取输入数据，读取器不一定需要提供从文件读取的数据

#### 跳过不良记录

用户代码的错误导致崩溃，修复错误 忽略(跳过)错误  检测记录确定性崩溃，发送信号给主程序，下次重新执行就能跳过记号

本地调试

状态信息：状态页面查看所有任务的进度

计数器工具：对各种事件记录，在任务函数中调用incre，定期传给master以更新

## GO

方法是带着接受者参数的函数，接受者需要是包内定义的类型。带指针参数的函数必须接受一个指针，但是以指针为接收者的方法被调用时，接收者既能为值也能为指针。反过来也是对的，接受值参数的函数必须接受值，但是接收者既能为值也能为指针。

chan ：默认情况下，发送和接收操作在另一端准备好之前都会阻塞。这使得 Go 程可以在没有显式的锁或竞态变量的情况下进行同步

go 函数名(参数列表)

### 接口

接口：定义为接口后，把其他类型的实例赋值给他，然后可以对应的不同方法，来实现多态

接口是一种抽象的类型，有点像泛型

```go
// 接口不关你是什么类型，只要你实现了这个方法
// 定义一个类型，一个抽象的类型，只要实现了say()就可以称为sayer类型
type sayer interface(){
    say(参数) 返回值列表
    say2()
    say3()
}
// 自定义类型，实现了say方法
type cat struct{}
func (c cat) say(){
    fmt.Println("喵喵喵")
}
type dog struct{}
func (d dog) say(){
    fmt.Println("汪汪汪")
}
type person struct{}
func (p person) say(){
    fmt.Println("嗷嗷嗷")
}
// 接口类型的入口，interface作为入参类型
func da(arg sayer){
    arg.say()
}
// 主函数
func main(){
    c1 := cat{}
    da(c1)
    d1 := dog{}
    da(d1)
    p1 := person{}
    da(p1)
}
```



## lab1

```shell
go run mrsequential.go wc.so pg*.tx // 之后的两项是传给mrsequential的命令行参数，分别是一个动态库和所有电子书

文件wc.go以及mrapps目录下的其它几个文件，都定义了名为map, reduce的函数，这两个函数在mrsequential.go中加载并调用。给mrsequential绑定不同的*.so文件，也就会加载不同的map, reduce函数。如此实现某种程度上的动态绑定。

我们的代码主要写在src/mr目录下的几个文件，这几个文件由src/main目录下两个文件mrcoordinator.go, mrworker.go调用。这两个文件的作用是启动进程、加载map, reduce动态库，并进入定义在src/mr目录下的主流程
```

### 1. RPC和chan实现数据交换

#### PRC

实现worker与cooridinator之间**rpc交互** call函数通过端口，调用rpcname（Coordinator.函数名），远程调用（Coordinator.函数名）这个函数，然后通过内存地址读取取回结果

目的 ： RPC 约等于PC

客户端有一个函数，服务器端有一个函数的实现 ； RPC保证远程调用，传递参数，然后返回结果

过程： 在 RPC 中，客户端和服务器端都有一个 stub。Stub 是指一份代理，它代表了本地对象与远程对象之间的一个代理，实现了客户端和服务器端的交互。当客户端调用 stub 上的方法时，stub 内部会将请求（包括函数的类型、参数等）封装成一条消息并发送给服务器端，（中间数据有序列化和反序列化）服务器端接收到请求并将执行结果返回给 stub，然后 stub 再将执行结果返回给客户端（中间数据有序列化和反序列化）。这个过程中，stub 将客户端和服务器端的交互细节封装了起来，让客户端和服务器端像本地调用一样进行 RPC 调用，从而降低了复杂度和使用难度，提高了开发效率。

在客户端和服务器端的 stub 中，通常包含了一些调用远程方法所需的信息，如方法名、方法参数等，这些信息可以通过特定的协议进行传输。在传输层面，RPC 中比较常见的协议有 TCP、HTTP 和 HTTPS 等。在使用这些协议进行远程方法调用时，stub 还可以负责序列化和反序列化请求和响应消息，这些消息通常使用类似 JSON、XML 或二进制编码形式进行表示。

```go
 // 如果不需要参数和返回值，最好设置为空的结构体，不建议设置nil
    var reply int // 初始值为0
    err := client.Call("Arith.Ping", &struct{}{}, &reply)
    if err != nil {
        panic(err)
    }
```



##### 用法例子

下面是一个使用 Go 实现 RPC 调用的例子：

**server.go**

```go
package main

import (
    "errors"
    "net"
    "net/rpc"
    "time"
)

type Args struct {
    A, B int
}

type Arith int

func (t *Arith) Multiply(args *Args, reply *int) error {
    time.Sleep(1 * time.Second) // 模拟耗时操作
    *reply = args.A * args.B
    return nil
}

func main() {
    arith := new(Arith)
    rpc.Register(arith)
    rpc.HandleHTTP()
    l, err := net.Listen("tcp", ":5555")
    if err != nil {
        panic(err)
    }
    defer l.Close()
    for {
        conn, err := l.Accept()
        if err != nil {
            continue
        }
        go rpc.ServeConn(conn)
    }
}
```

这里定义了一个 `Arith` 类型，并为它实现了一个 `Multiply` 方法。`Multiply` 方法接收 `Args` 类型的参数，将其相乘后保存在 `reply` 指针指向的整型变量中。

程序启动后，会监听在 5555 端口，并等待客户端的 RPC 调用请求。

**client.go**

```go
package main

import (
    "fmt"
    "net/rpc"
    "time"
)

type Args struct {
    A, B int
}

func main() {
    client, err := rpc.DialHTTP("tcp", "localhost:5555")
    if err != nil {
        panic(err)
    }
    defer client.Close()

    args := &Args{7, 8}
    var reply int
    start := time.Now()
    err = client.Call("Arith.Multiply", args, &reply) // call stub生成消息 远程调用 

    if err != nil {
        panic(err)
    }
    fmt.Printf("Arith: %d*%d=%d,time: %v\n", args.A, args.B, reply, time.Since(start))
}
```

这里定义了一个Args结构体类型, 使用rpc.DialHTTP来与RPC服务端建立连接, 然后向服务端发送一个 Multiply 方法的 RPC 调用，入参是 args，返回值是 reply。最后，将服务端返回的结果打印出来。

当程序运行时，客户端会向服务端发送一个 args.A=7,args.B=8 的请求，服务端在收到请求后会执行 Multiply 方法，将返回值 56 填充到 reply 变量中并返回给客户端。最后客户端打印出这个结果。

当服务端执行 Multiply 方法时，还用了一条 `time.Sleep(1 * time.Second)` 语句来模拟一个耗时操作。这样，我们就可以看到整个 RPC 调用过程的耗时了。



#### chan

channel 是golang特有的类型化消息的队列，可以通过它们发送类型化的数据在协程之间通信，可以避开所有内存共享导致的坑；通道的通信方式保证了同步性。数据通过通道：同一时间只有一个协程可以访问数据：所以不会出现数据竞争，设计如此。数据的归属（可以读写数据的能力）被传递。

```go
var ch chan Type // 定义类型为 Type 的 chan
ch <- data // 发送数据 data 到 chan ch
data := <- ch // 从 chan ch 中接收数据，并将其赋值给变量 data   接收操作会一直阻塞，直到有数据可以被接收。如果不希望接收操作被阻塞，可以使用带缓冲的 chan
var ch = make(chan Type, capacity) // 带有缓冲区的 chan 容量为 capacity  capacity 表示 chan 的缓冲区大小，可以在创建 chan 的时候指定。带缓冲的 chan 可以在没有读取者时发送数据，也可以在没有写入者时接收数据，但是如果缓冲区已满或为空时，发送和接收操作仍然会被阻塞

```

此外，chan 还可以使用 select 语句实现多路复用，同时处理多个 chan 的读写操作，类似于 Linux 中的 select() 系统调用。在 select 语句中，可以通过 case 语句处理 chan 的读写操作，default 语句用于处理没有可读和可写的 chan 的情况。当有多个 case 同时满足条件时，程序会随机选择一个 case 进行处理。例如：

```go
select {
  case data := <- ch1:
    // 处理从 ch1 中接收到的数据
  case ch2 <- data:
    // 处理将数据 data 发送到 ch2 中
  default:
    // 没有任何 chan 可读也没有可写的数据
}
```



##### 用法例子

下面是一个简单的 Go 程序，使用 chan 实现了两个协程之间的通信。在这个程序中，有一个 jobs chan 用于发送工作任务，一个 results chan 用于接收工作结果。有3个协程(worker)同时从 jobs chan 中接收任务，完成任务后将结果发送到 results chan 中。在 main 函数中，我们先向 jobs chan 发送了5个任务，然后通过接收 results chan 中的结果来等待工作的完成。这样，就实现了两个协程之间简单的通信和同步。

```go
package main

import (
    "fmt"
    "time"
)

func worker(id int, jobs <-chan int, results chan<- int) {
    // <-chan是receive-only channel而 chan<-是 send-only channel
    // 在 Go 中，如果将一个 chan 赋值给另一个 chan 变量时，就会复制一个新的 chan,而且读写的属性一致
    // 在 Go 中，一个 chan 可以同时被多个协程读或写，但会有竞争写入或者竞争读取的问题，需要通过锁或者其他同步原语来解决。为了保证并发安全，建议在使用 chan 时，尽量将读写 chan 的操作限制在一个协程中，每个 chan 只由一个协程读或写。
    for j := range jobs {
        fmt.Printf("worker %d started job %d\n", id, j)
        time.Sleep(time.Second)
        fmt.Printf("worker %d finished job %d\n", id, j)
        results <- j * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    // 启动三个协程进行工作
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // 发送 5 个任务到 jobs chan 中
    for i := 1; i <= 5; i++ {
        jobs <- i // 主协程写一个，协程池就会拿一个，chan并发安全 保证只有一个协程拿到运行，是线程安全的
    }
    close(jobs)

    // 接收工作结果
    for r := 1; r <= 5; r++ {
        <-results
    }
}
```

1. 我的做法是涉及的worker和cooridinator交互都用RPC交互，（其实部分用chan也可以吧）；
2. task相关信息队列的用chan就可以，用来分发task，天然安全，不会race

### 2. map和reduce处理

map 处理，用ihash处理下key分成Nreduce份用json编码后写出到"mr-x-y"文件。注意mr论文这步是有排序的，因为真正生产活动数据量是非常巨大的，map端提前排序好后，reduce的排序压力会减小很多。这里排不排序无所谓

map 处理结果 返回kv，存储在中间文件中，中间文件命名 mr-X-Y X是map任务号，y是reduce任务号；之后reduce再都读取中间文件；

因为真正分布式worker都不在一个机器上，涉及**网络传输**，所以用**json编码解码走个过场**

#### 设计结果：

map阶段应该将中间键划分为**nReduce**的存储桶，这里只用了一倍reduce数量也就是**%NumReduce**；每个映射器都应创建NumReduce个中间文件，存在main目录下以供Reduce任务使用。

Worker实现应该将第X个Reduce任务的输出放入文件MR-OUT-X中。Done()方法，在MapReduceJob完全完成时返回TRUE；此时，mrcoherator.go将退出。当作业完全完成时，工作进程应该退出。

#### Json编码

纯字符串形式的数据，轻量级的数据交换格式，可以跨平台、跨语言使用。

[JSON（JavaScript Object Notation）是一种轻量级的数据交换格式，它是基于 JavaScript 的一个子集，采用完全独立于编程语言的格式来表示数据，可以跨语言、跨平台使用。简洁清晰的层次结构使得 JSON 逐渐替代了 XML，成为了最理想的数据交换格式，广泛应用于 Web 开发领域](http://c.biancheng.net/json/what-is-json.html)[1](http://c.biancheng.net/json/what-is-json.html)[2](http://www.json.org/json-zh.html)。

JOSN编码和解码

```go
				enc := json.NewEncoder(tmp_file) // 创建JSON编辑器，用于写入tmp_file
				err = enc.Encode(bucket[i])     // 将buckt[i]的值编码为JSON格式，写入文件
```

```go
				for {
					var kv []KeyValue
					if err := dec.Decode(&kv); err != nil { //解码
						break
					}
					intermediate = append(intermediate, kv...) // ...表示把切片打散成元素后加入
				}
```

### 3 加锁

由于我用了channel所以在任务分配队列实现了天然并发安全，但是在别的地方还是遇到了问题，比如Done函数通过mrcoordinator主线程去时不时读取coordinator的状态来判断是否结束死循环。还有在一个worker调coordinator拉取数据的时候，另一个worker调coordinator的checkJobDone()函数进行检查。因此在响应可能发生冲突的地方加锁

### 4. 等待

1. GetTask调用失败，原因未处理work等待；
   worker有时候需要等待,比如当map任务都分发出去了,有的worker完成后又来申请任务，此时还有map未完成,reduce不能开始，这个worker需要等待下

解决：每次worker工作后都time.Sleep(time.Second)

### 5.crash test

1.  crash test失败，原因未处理可能损坏的worker
   mrapps/crash.go 随机干掉map reduce,看crash.go的代码是有一定几率让worker直接退出或者长时间延迟,可以用来测试恢复功能。这个逻辑是整合在map reduce函数里面的，注意worker被干掉时候任务已经拿到手了

解决：添加TimeTick函数，放到Done()函数中，检查map task 和 reduce task的时间，如果超过10秒，重新加载该Task

```go
				c.MapTask <- Task{FileName: c.files[i], MapId: i} //TODO 这里硬加Task没有去掉原来的Task 可能会导致跑很多次
/* task生命周期分析： 
		maptask在MakeCoordinator创建，定时间； 在做完在TaskFin设置true ； 在TimeTick检查是否超时，如果超时 重新添加这个任务
		这里其实判断超时没有区分是创建时间超时还是做任务超时，所以在做任务的时候设置时间
```

要记住go 的range第一个参数是index 

 strconv.Itoa 转化格式 字符串

os.Open 默认以只读方式打开，返回一个error； os.OpenFile可以设置打开文件的模式

## GFS

### 高容错的实现难点： 

1. 高性能的要求 -> 跨机器对数据分片
2. 大量计算机   ->  会有很多崩溃
3. 容错解决奔溃情况 -> 副本 
4. 副本 -> 需要实施更新，保持副本一致
5. 一致性高 -> 性能开销

### 一致性实现的难点 ：

1. 并发性 ，保证线程安全，可以通过之后的分布式锁等机制解决
2. 故障和失败案例 ：为了容错会用复制，但是不成熟的复制会导致读者在不做修改的情况下读到两次不同数据。

### GFS简介

 GFS旨在保持高性能，且有复制、容错机制，但很难保持一致性，现在谷歌已经不用了； 在论文中可以看到mapper从GFS系统(上千个磁盘)能够以超过10000MB/s的速度读取数据。论文发表时，当时单个磁盘的读取速度大概是30MB/s，一般在几十MB/s左右。

GFS的几个主要特征：

- Big：large data set，巨大的数据集
- Fast：automatic sharding，自动分片到多个磁盘
- Gloal：all apps see same files，所有应用程序从GFS读取数据时看到相同的文件（一致性）
- Fault tolerance：automic，尽可能自动地采取一些容错恢复操作

### GFS的Master的工作

- 维护文件名到块句柄数组的映射(file name => chunk handles)  **filename -> chunk handles -> chunk server \  version**

  这些信息大多数存放在内存中，所以Master可以快速响应客户端Client

- 维护每个块句柄(chunk handle)的版本(version)

- 维护块存储服务器列表(list of chunk servers)

  - 主服务器(primary)
    - Master还需维护每一个主服务器(primary)的**租赁时间(lease time)**
  - 次要服务器(secondaries)

  典型配置即将chunk存储到3台服务器上

- log+check point：通过日志和检查点机制维护文件系统。所有变更操作会先在log中记录，后续才响应Client。这样即使Master崩溃/故障，重启时也能通过log恢复状态。master会定期创建自己状态的检查点，落到持久性存储上，重启/恢复状态时只需重放log中最后一个check point检查点之后的所有操作，所以恢复也很快。

 这里需要思考的是，哪些数据需要放到稳定的存储中(比如磁盘)？

- 比如file name => chunk hanles的映射，平时已经在内存中存储了，还有必要存在稳定的存储中吗？

  需要，否则崩溃后恢复时，内存数据丢失，master无法索引某个具体的文件，相当于丢失了文件。

- chunk handle 到 存放chunk的服务器列表，这一层映射关系，master需要稳定存储吗？

  不需要，master重启时会要求其他存储chunk数据的服务器说明自己维护的chunk handles数据。这里master只需要内存中维护即可。同样的，主服务器(primary)、次要服务器(secondaries)、主服务器(primary)的租赁时间(lease time)也都只需要在内存中即可。

- chunk handle的version版本号信息呢，master需要稳定存储吗？

  需要。否则master崩溃重启时，master无法区分哪些chunk server存储的chunk最新的。比如可能有服务器存储的chunk version是14，由于网络问题，该服务器还没有拿到最新version 15的数据，master必须能够区分哪些server有最新version的chunk。

### GFS数据读取流程

![image-20230411161308484](D:\MyTxt\typoraPhoto\image-20230411161308484.png)

GFS通过Master管理文件系统的元数据等信息，其他Client只能往GFS写入或读取数据。

应用并发的通过GFS Client读取数据时，单个读取的大致流程如下：

1. Client向Master发起读数据请求
2. Master查询需要读取的数据对应的目录等信息，汇总文件块访问句柄、这些文件块所在的服务器节点信息给Client（大文件通常被拆分成多个块Chunk存放到不同服务器上，单个Chunk很大， 这里是64MB）
3. Client得知需要读取的Chunk的信息后，直接和拥有这些Chunk的服务器网络通信传输Chunks

### GFS文件读取

1. Client向Master发请求，要求读取X文件的Y偏移量的数据
2. Master回复Client，X文件Y偏移量相关的块句柄、块服务器列表、版本号(chunk handle, list of chunk servers, version)
3. Client 缓存cache块服务器列表(list of chunk servers)  
   ？ 只有一个Master，所以需要尽量减少Client和Server之间的通信次数，缓冲减少交互
4. Client从最近的服务器请求chunk数据(reads from closest servers)
   ？ 因为这样在宛如拓扑结构的网络中可以最大限度地减少网络流量(mininize network traffic)，提高整体系统的吞吐量
5. 被Client访问的chunk server检查version，version正确则返回数据
   ？ 为了尽量避免客户端读到过时数据的情况。

### GFS文件写入

 主要是关注文件写入的append操作，在mapreduce 中，reduce处理后计算结果需要append到file中

<img src="D:\MyTxt\typoraPhoto\image-20230411165609018.png" alt="image-20230411165609018" style="zoom: 67%;" />

1. Client向Master发出请求，查询应该往哪里写入filename对应的文件。

2. Master查询filename到chunk handle映射关系的表，找到需要修改的chunk handle后，再查询chunk handle到**chunk server数组映射关系的表**，以list of chunk servers(主副本primary、secondaries、version信息)**也就是得到所有该文件的chunkserver并且附加上版本号**作为Client请求的响应结果

   接下去有两种情况，已有primary和没有primary(假设这是系统刚启动后不久，还没有primary)

   - 有primary，（**primary：主文件副本，也就是正常情况下对外读写的文件；secondaries： 其他副本**）

     继续后续流程

   - 无primary

     - master在chunk servers中选出一个作为primary，其余的chunk server作为secondaries

       (暂时不考虑选出的细节和步骤)

       - master会增加version（每次有**新的primary**时，都需要考虑时进入了一个new epoch，所以需要**维护新的version**），然后向primary和secondaries发送新的version，并且会发给primary有效期限的租约lease。这里primary和secondaries需要将version存储到磁盘，否则重启后内存数据丢失，无法让master信服自己拥有最新version的数据(同理Master也是将version存储在磁盘中)。

3. Client发送数据到想写入的chunk servers(primary和secondaries)，有趣的是，**这里Client只需访问最近的secondary，而这个被访问的secondary会将数据也转发到列表中的下一个chunk server**，**此时数据还不会真正被chunk severs存储**。（即上图中间黑色粗箭头，secondary收到数据后，马上将数据推送到其他本次需要写的chunk server）
   ？ 这么做提高了Client的吞吐量，避免Client本身需要消耗大量网络接口资源往primary和多个secondaries都发送数据。

4. 数据传递完毕后，Client向primary发送一个message，表明本次为append操作。

   primary此时需要做几件事：

   1. primary此时会检查version，如果version不匹配，那么Client的操作会被拒绝
   2. primary检查**lease是否还有效**，如果自己的lease无效了，则不再接受任何mutation operations（因为租约无效时，外部可能已经存在一个新的primary了）
   3. 如果version、lease都有效，那么primary会选择一个offset用于写入
   4. primary将前面接收到的数据写入稳定存储中

5. primary发送消息到secondaries，表示需要将之前接收的数据写入指定的offset（更新其他副本）

6. secondaries写入数据到primary指定的offset中，并回应primary已完成数据写入

7. primary回应Client，你想append追加的数据已完成写入

### GFS一致性

1. appen一次之后，read返回什么结果

   例如： M（maseter），P（primary），S（Secondary）

   某时刻Mping不到P了，那咋办？ 

   假设选出新P，那可能旧P还在和CLient交互，两个P同时存在叫做脑裂，会出现严重问题，比如数据写入顺序混乱等问题，严重的脑裂问题可能会导致系统最后出现两个Master。

   所以依照M知道的旧P的lease，在旧Please期限结束之前，不会选出新P，也就是M和P无法通信，但是P还可能和其他Client通信

2. 强一致性的保证
   保证 所有S都成功写入或者都不写入

GFS是为了运行mapreduce而设计的

## VMWare-FT

### The Design of a Practical System for Fault-Tolerant Virtual Machines

(容错虚拟机)[[https://blog.csdn.net/weixin_40910614/article/details/117995014]]

只针对一个backup 单核 一个storage

### 备份实现方法

实现分布式容错常用方法是 **主/备份方法**；一致性的保证需要保持同步，两种方法： 

1. 状态转移，主服务器持续发送搜索状态变化给备份服务器，带宽需求大
2. 备份状态机，备份状态机视为确定状态机（DSM）。主/备份服务器相同的初始状态启动，然后相同的顺序接受并执行相同的输入请求，得到相同的输出，这样保持一致性，方法复杂，数据传输量小很多。
3. 状态转移是记录checkpoint点，同步完后再返回给client；复制状态机是操作记录同步后，再响应client

采用备份状态机会有一个问题，即对于真实物理机来讲，有些行为是不确定的（例如，中断，产生的随机数等），所以很难将行为看作是一个确定性（deterministic）的行为。
VMWare的解决方案是将所有操作虚拟化，并在VM和物理机之间加一个Hypervisor（虚拟机管理程序），然后就可以通过hypervisor让VM都能模拟成一个DSM运行，那么primary的所有操作都确定了，再通过Logging channel将这些操作传递到backup，然后backup进行replay，从而实现两者的高度一致。
具体是怎么将不确定的行为进行确定化操作的，它主要是进行了一个写日志的操作来保证的

### 复制状态机具体复制的是什么操作

应用程序级别： 比如GFS的文件append或write，需要关注复制状态机如何处理的，一般需要修改应用程序本身

机器层面操作： 状态state是寄存器的状态、内存的状态，操作operation则是传统的计算机指令；这种情况复制状态机感知的是底层机器指令，用硬件很昂贵，这里用的是虚拟机实现的

<img src="D:\MyTxt\typoraPhoto\image-20230413151512850.png" alt="image-20230413151512850" style="zoom: 50%;" />

### 非确定信息导致差异

非确定性的指令，网络包接受/时钟中断，多核（不讨论）

VM-FT对非确定性指令会先拦截，然后模拟执行，并且同步

### 失败场景处理

primary操作完成后，和client断了，这时候backup接手，那这次操作就失效了。

实际上不会发生这个错误，primary响应client之前，会通过logging channel发送消息给backup，backup相同操作后发ack，primary收到ack后，才响应client

raft也是一样

### 基本架构

1. 避免脑裂，对外只有一个primary输入输出，通过logging channel传递给Backup达到同步
2. 两个虚拟机部署，一个磁盘，一个网络
3. 确定性存放：也就是做一个确定状态机，具体是**Hypervisor**操作虚拟化后，把所有primary的确定、不确定（比如随机生成一个数）事件都记录log entry流，传给backup，保证一致性
4. FT Protocol：每次primary把output**写到日志给backup后才发送**，目的是保证backup随时能接手且**一致**
   保证**故障转移**（当primary失败时，我们希望切换到backup提供服务）
5. FT Logging Buffers and Channel： primary和backup两端的日志缓存差距保持在一个范围，用到了生产者消费者模型
6. 检测 故障响应：primary和backup两端的UDP心跳检测、logging channel检测
7. 虚拟机恢复FT VMotion：对VM进行克隆

### 与GFS比较

1. VM-FT备份的是**计算**，GFS备份的是存储。
2. VM-FT提供了相对严谨的一致性，并且对客户端和服务器都是透明的。可以用它为任何已有的网络服务器提供容错。例如，可以将 FT 应用于一个已有的邮件服务器并为其提供容错性。
3. GFS 只针对一种简单的服务提供容错性，它的备份策略会比 FT 更为高效。例如，GFS 不需要让所有的操作都应用在所有的 Replica 上。
4. VM-FT**不适合进行高吞吐量的服务**，性能下降很多，毕竟它要对整个系统内的内存、磁盘、中断等都进行 replay，这是一个巨大的负担

## Raft

### 复制状态机

复制状态机（Replicated State Machine，简称RSM）是一种分布式系统的设计模式。在该模式下，一个服务或应用程序的状态机被复制到多个节点上并进行并行处理，以提高可用性和性能。

具体地说，采用复制状态机的系统中，对于某个特定的客户端请求，它将被发送到所有复制的状态机或者其中的一组。然后，每个状态机独立地执行相同的操作序列，并生成相同的结果。最终，生成的结果将会被汇总并返回给客户端。

通过这种方式，复制状态机可以将单点故障风险降至最低并提高系统的可靠性。此外，由于并行执行相同的操作序列，该模式还能够提供更好的性能和可伸缩性。

复制状态机是一种常见的分布式系统设计模式，在诸如Google、Facebook和Amazon等互联网巨头公司的分布式系统中得到广泛应用。

### 单点故障

前面介绍过的复制系统，都存在单点故障问题(single point of failure)。

- mapreduce中的cordinator
- GFS的master
- VM-FT的test-and-set存储服务器storage

 而上述的方案中，采用单机管理而不是采用多实例/多机器的原因，是为了避免**脑裂(split-brain)**问题。

 不过大多数情况下，单点故障是可以接受的，因为单机故障率显著比多机出现一台故障的概率低，并且重启单机以恢复工作的成本也相对较低，只需要容忍一小段时间的重启恢复工作。

**为什么单机管理能避免脑裂问题** ？

比如有两个strorage，要选出primary，那可能有网络分区的原因，storage两个分区产生两个primary，对外界的client来说有两种不同的

### 大多数原则 majority rule

如何解决脑裂： 如果投票得大于一半，多数的那个成为leader，也就是多数的分区会继续运行，如果没有多数 系统不能运行

### 用Raft构造复制状态机RSM

这里raft就像一个library应用包。假设我们通过raft协议构造了一个由3台机器组成的K/V存储系统。

系统正常工作时，大致流程如下：

- Client向3台机器中作为leader的机器发查询请求
- leader机器将接收到的请求记录到底层raft的顺序log中
- 当前leader的raft将顺序log中尾部新增的log记录通过网络同步到其他2台机器
- 其他两台K/V机器的raft成功追加log记录到自己的顺序log中后，回应leader一个ACK，
  （如果有一台网络问题比较慢，那也达到了majority）
- leader的raft得知其他机器成功将log存储到各自的storage后，将log操作反映给自己的K/V应用
  （网慢的那台这时候才发ACK，）
- K/V应用实际进行K/V查询，并且将结果响应给Client

系统出现异常时，发生如下事件：

- Client向leader请求
- leader向其他2台机器同步log并且获得ACK
- leader准备响应时突然宕机，无法响应Client
- 其他2台机器重新选举出其中1台作为新的leader
- Client请求超时或失败，重新发起请求，**系统内部failover故障转移**，所以这次Client请求到的是新leader
- 新leader同样记录log并且同步log到另一台机器获取到ACK
- 新leader响应Client

### Raft的log

用途：持久化、顺序化 操作数据 ，方便重传，方面查看同步操作进行的情况

格式：

有很多的log entry（入口），比如log index 、 leader term这些唯一标识；每个log entry 有 command 和 leader term信息；

保证最后机器的所有log的一致性

### Raft的选举

followers收不到leaders的心跳，election time超时，开始重新选举，新的leader term

这时候就算原leader接进来，也会发现是新的leader term， 不会唱脑裂

#### 防止选举死循环

如果两个followers的election time 几乎同时到齐，都成为candidate，那会一直竞争leader死循环;

一般采用election time为随机值，防止同时发起选举

#### 选举超时时间

略大于心跳时间，太短会频繁选举，而且选举过程中是对外宕机，太长检测不到，本身也是宕机

加入些随机数，防止分裂选举死循环

 Raft论文进行了大量实验，以得到250ms～300ms这个在它们系统中的合理值作为eleciton timeout。

## FT-Raft-2

### 选举规则

处理别节点发来的RequestVote RPC时，需要检查限制，满足以下之一才能赞同票：

1. 候选人最后一条Log条目的任期号**大于**本地最后一条Log条目的任期号；或者
2. 候选人最后一条Log条目的任期号**等于**本地最后一条Log条目的任期号，且候选人的Log记录长度**大于等于**本地Log记录的长度

成为leader的限制：

1. 大多数原则
2. 当选的机器一定是具有最新的term的机器

### 日志覆写同步（未优化版本）

所有raft节点维护两个Index

* nextIndex数组：乐观的变量，所有raft节点都维护`nextIndex[followerId]`用于记录leader认为followerId的下一个需要填充log的index。
  更新时间：每个leader当选之后都会乐观的认为所有的follower的nextIndex是自己的log最后的下一个，而在实际appendEntries或者installSnapshot的时候如果发现日志同步没有那么乐观就会根据情况减小next，把之前没有同步的log先同步上。

* matchIndex数组：悲观变量，，所有raft节点都维护`matchIndex[followerId]`用于记录leader认为followerId的已经确认的log最后一个的index，表示在此之前的log都是和当前的leader确认一致的。
  更新时间：每个leader当选之后都会悲观的认为自己已经确认过所有的follower的nextIndex是0，因为还没开始确认。

未优化版本的问题：Raft集群中有出现log落后很多的server，leader需要进行很多次请求才能将其log与自己对齐

### 日志擦除

新上任的leader发现其他follower有和自己不一样的log，就会把从那个log开始往后的logs都擦除，来保证leader的日志权威性一致性

### 日志快速覆写同步（优化）

落后较多的log（比如新机器接入、宕机恢复很久）要同步到现在leader一致的话，按照之前的逐步回退很慢，浪费网络资源

优化后的log catch up quickly过程：

|      | logIndex1 | logIndex2 | logIndex3 | logIndex4 | logIndex5 |
| ---- | --------- | --------- | --------- | --------- | --------- |
| S1   | term4     | 5         | 5         | 5         | 5         |
| S2   | term4     | 6         | 6         | 6         | 6         |

1. S2的leader是term7当选，那nextIndex = 6，发送hearbeat随带log是(空, 6, 5)，意思是(当前nextIndex指向的term，nextIndex-1的term, nextIndex-1的值)
2. S1收到心跳，对比自己的logIndex为term5，与之前不同的是，除了no顺带回复自己的log信息(5,2), 意思是(请求中logIndex位置的值，当前值最早出现的logIndex位置)
3. S2收到回应后，把nextIndex改为2，下次附带([6,6,6,6],4,1), 意思是nextIndex之后的数据是[6,6,6,6]
4. S1收到后，检查logIndex1是term4对齐了，更新一致性

广播的时候，面对大量followers， 用每个go程单独处理followers

### 持久化

就是把一些全局的变量（比如currentTerm）写到（持久化存储）磁盘里，当然操作回复之前写入，类似于同步过程

持久化原因： 如果重启之后不知道任期号，很难确保任期只有一个leader

一个Raft节点崩溃重启（和新加入节点一样）后，必须重新加入

除了重新加入重新执行本地的log，更偏向于快速重启，上次持久化快照位置开始，这就要考虑持久化一些状态量：

- vote for：投票情况，因为需要保证每轮term每个server只能投票一次
- log：崩溃前的log记录，**因为我们需要保证(promise)已发生的(commit)不会被回退**。否则崩溃重启后，可能发生一些奇怪的事情，比如client先前的请求又重新生效一次，导致某个K/V被覆盖成旧值之类的。
- current term：崩溃前的当前term值。因为选举(election)需要用到，用于投票和拉票流程，并且**需要保证单调递增**(monotonic increasing)
- lastIncludedIndex: 奔溃前的快照保存的最后一个log的index，到这个log位置都是保存在快照中的，也就是状态机的持久化数据，如果崩溃后重启就不用交这个log之前的log了，因为上层可以直接读取整个快照。
- lasIncludedIndex: 奔溃前的快照保存的最后一个log的term

### 服务恢复

类似的，服务重启恢复时有两种策略：

1. 日志重放(replay log)：理论上将log中的记录全部重放一遍，能得到和之前一致的工作状态。这一般来说是很昂贵的策略，特别是工作数年的服务，从头开始执行一遍log，耗时难以估量。所以一般人们不会考虑策略1。
2. **周期性快照(periodic snapshots)**：假设在i的位置创建了快照，那么可以裁剪log，只保留i往后的log。此时重启后可以通过snapshot快照先快速恢复到某个时刻的状态，然后后续可以再通过log catch up或其他手段，将log同步到最新状态。（一般来说周期性的快照不会落后最新版本太多，所以恢复工作要少得多）

 这里可以扩展考虑一些场景，比如Raft集群中加入新的follower时，可以让leader将自己的snapshot传递给follower，帮助follower快速同步到近期的状态，尽管可能还是有些落后最新版本，但是根据后续log catch up等机制可以帮助follower随后快速跟进到最新版本log。

 使用快照时，需要注意几点：

- 需要拒绝旧版本的快照：有可能收到的snapshot比当前服务状态还老
- 需要保持快照后的log数据：在加载快照时，如果有新log产生，需要保证加载快照后这些新产生的log能够能到保留

### 使用Raft

 重新回顾一下服务使用Raft的大致流程

1. 应用程序中集成Raft相关的library包
2. 应用程序接收Client请求
3. 应用程序调用Raft的start函数/方法
4. 下层Raft进行log同步等流程
5. Raft通过apply channel向上层应用反应执行完成
6. 应用程序响应Client

- 并且前面提过，可能作为leader的Raft所在服务器宕机，所以Client必须维护server列表来切换请求的目标server为新的leader服务器。
- 同时，有时候请求会失败，或者Raft底层失败，导致重复请求，而我们需要有手段辨别重复的请求。通常可以在get、put请求上加上请求id或其他标识来区分每个请求。一般维护这些请求id的服务，被称为clerk。提供服务的应用程序通过clerk维护每个请求对应的id，以及一些集群信息。

## Lab2

<img src="D:\MyTxt\typoraPhoto\image-20230508170955517.png" alt="image-20230508170955517" style="zoom:33%;" />

![image-20230508171059069](D:\MyTxt\typoraPhoto\image-20230508171059069.png)

Raft 复制状态机协议

实现FT ，难点是故障导致副本一致性受损，难点是调试

把client请求组织到log序列汇总，确保所有副本服务器看到相同的Log，来保证一致性；如果服务器故障之后回复，Raft会将日志更新；所以只要大多数服务器活着且能通信，Raft就继续，不然就宕机直到大多数机器能通信才开始从中断位置通信

Raft作为关联方法的Go对象类型，作为模块。一组RAFT实例通过RPC相互通信以维护复制的日志。您的RAFT界面将支持不确定的编号命令序列，也称为日志条目。这些条目用索引号进行编号。具有给定索引的日志条目最终将被提交。此时，您的RAFT应该将日志条目发送到更大的服务以供其执行。

### 时间

三个时间的比较： broadcastTime 100« electionTimeout 200-400« 平均故障时间

两个RPC ：  AppendEntriesRPCs是 leader进行日志复制和心跳时使用的 << RequestVoteRPCs是候选在这选举过程中使用的

两个时间驱动用单独的goroutine驱动 

简单的实现方式是在raft中维护一个变量，包含leader的最后一次消息的时间，并用goroutine用sleep()驱动定期检查是否超时

定时的设置想了很久，把elctiontime作为下一个选举的时间点，定期检查是否超过这个时间点，如果超过了就要选举，heartbeattime设置为时间长度，用sleep相应长度然后发送heartbeat	

先设置 100 和 150-300，一次心跳不到重选

### log

- `entry`：Raft 中，将每一个事件都称为一个 entry，每一个 entry 都有一个表明它在 log 中位置的 index（之所以从 1 开始是为了方便 `prevLogIndex` 从 0 开始）。只有 leader 可以创建 entry。entry 的内容为 `<term, index, cmd>`，其中 cmd 是可以应用到状态机的操作。在 raft 组大部分节点都接收这条 entry 后，entry 可以被称为是 committed 的。
- `logs`：由 entry 构成的数组，只有 leader 可以改变其他节点的 log。 entry 总是先被 leader 添加进本地的 log 数组中去，然后才发起共识请求，获得 quorum 同意后才会被 leader 提交给状态机。follower 只能从 leader 获取新日志和当前的 commitIndex，然后应用对应的 entry 到自己的状态机

### election

candida并发请求选票，大于一半的时候成为leader，保证只成为一次 用sync.Once类型

这是Go编程语言中的代码，`sync.Once`是Go标准库中的一种并发原语，用于确保某个函数只会被执行一次。

在该代码中，`becomeLeader`是一个变量，使用`sync.Once`作为其类型，在程序运行过程中，我们可以通过调用`becomeLeader.Do(func)`的方式来确保`func`只会被执行一次。每当我们需要执行`func`时，我们可以多次调用`becomeLeader.Do(func)`函数，但实际上，该函数只会执行一次。

### RPC的处理

每个RPC应该在自己的goroutine中发送并处理回复，因为：

1. 到不了的peer不会延迟收集选票的过程
2. electiontime和heartbeattime可以继续在任何时候计时

election是并行的给集群中的其它机器发送 RequestVote RPCs

选取投票的写法： 一个函数开始投票，分出很多go程每个单独对接，每个函数对接发RPC和后处理

### RequestVote

安全性的限制：Raft使用选举过程来保证一个候选者必须包含有所有已提交的日志才能胜出；请求中包含了 leader的日志信息，如
果投票者的日志比候选者的日志更新，那么它就拒绝投票 :

1. 候选人最后一条Log条目的任期号**大于**本地最后一条Log条目的任期号；或者
2. 候选人最后一条Log条目的任期号**等于**本地最后一条Log条目的任期号，且候选人的Log记录长度**大于等于**本地Log记录的长度

成为leader的限制：

1. 大多数原则
2. 当选的机器一定是具有最新的term的机器

### AppendEntries 

每次当 leader发送 AppendEntries RPCs请求的时候，请求中会包含当前nextIndex后面的日志记录 和 他直接前继的任期和索引，

1. 如果存在一条日志索引和 prevLogIndex相等，但是任期和 prevLogItem不相同的日志，需要删除这条日志及所有后继日志。

2. 如果 leader复制的日志本地没有，则直接追加存储。

以上两条需要分别进行，如果直接用leader发来的日志记录覆盖follower的日志，那会产生bug，因为这个RPC可能是过时的RPC，所以需要严格保存这两条的分别执行。

### applyCh

作为一个ApplyMsg的chan，日志提交之后 需要添加发送的条目

需要用单独的goroutine实现，因为可能阻塞，必须是一个不然难以确保按照日志顺序发送

go程用sync.Cond在不满足发送条件的时候等待, 需要提交的时候唤醒

### 日志同步

nextIndex是leader认为的下次发给其他follower新log的首位置，在当选的时候会自认为所有follower的nextIndex都和自己一样，是自己日志的记录的最后一条的+1位置；

commitIndex是每个server被提交日志后最新log的索引。

lastApplied是每个server提交状态机的最新log的索引。

**Raft 保证下列两个性质**：

- 如果在两个日志（节点）里，有两个 entry 拥有相同的 index 和 term，那么它们一定有相同的 cmd；
- 如果在两个日志（节点）里，有两个 entry 拥有相同的 index 和 term，那么它们前面的 entry 也一定相同。

通过”仅有 leader 可以生成 entry”来确保第一个性质， 第二个性质则通过一致性检查（consistency check）来保证，该检查包含几个步骤：

leader 在通过 AppendEntriesRPC 和 follower 通讯时，会带上上一块 entry 的信息， 而 follower 在收到后会对比自己的日志，如果发现这个 entry 的信息（index、term）和自己日志内的不符合，则会拒绝该请求。一旦 leader 发现有 follower 拒绝了请求，则会与该 follower 再进行一轮一致性检查， 找到双方最大的共识点，然后用 leader 的 entries 记录覆盖 follower 所有在最大共识点之后的数据。

寻找共识点时，leader 还是通过 AppendEntriesRPC 和 follower 进行一致性检查， 方法是发送再上一块的 entry， 如果 follower 依然拒绝，则 leader 再尝试发送更前面的一块，直到找到双方的共识点。 因为分歧发生的概率较低，而且一般很快能够得到纠正，所以这里的逐块确认一般不会造成性能问题。当然，在这里进行二分查找或者某些规则的查找可能也能够在理论上得到收益。

#### 边界条件考虑

初始存入 0 0 空作为第一个log，作为dummy节点。在添加快照功能之后需要把logs[0].index改成lastIncludedIndex。在这之前，保持dummy节点作为锚点不被改变

#### 快速同步

实现的主要方法是在reply中添加一个Xindex，每次RPC结束后都通知leader更新nextIndex[serverID]。

如果落后的较多，则返回Xindex为最后一个log的index + 1通知leader发送下次从这个Xindex开始日志串

如果index匹配了，但是term不一致，则按照论文给出的优化方法，找到这个term对应的第一个日志并放到Xindex中存起来，这样 leader
接受到响应后，就可以直接跳过所有冲突的日志（其中可能包含了一致的日志）。这样就可以减少寻找一致点的过程。

按照这个方案2B的catch up测试还是比较慢，用27秒，老师给出的是17秒

#### 一个RPC重试的BUG

原来的实现：

​	如果appendEntriesLeader发现follower的日志不同步，重新修改后递归appendEntriesLeader，也就是重发RPC直到同步完成。但是因为本身就是上锁的函数，递归调用自己，还要解锁再加锁，就算这样了递归调用不知道有错，错误不匹配后retry，但是第二次还出错不retry了。

修改后的实现：

​	如果同步失败，则修改nextIndex，等着下一次心跳或者快照来同步

### 测试结果说明

| 时间 | peer数量 | RPC数量 | PRC总字节数 | RPC中日志目录条数 |
| ---- | -------- | ------- | ----------- | ----------------- |
| 3.1  | 3        | 60      | 15190       | 0                 |



### 日志提交一致性

leader知道有log entry到达majority servers，他就会commit这个log。

有特殊情况是达到了majority servers但还没commit的leader crash，所以不通过判断majority servers来判断是不是该commit，因为就算达到了majority server之后的leader可能还会全部覆盖。Raft采用只有的currentTerm的leaders的logs被提交，因为选主过程中的**Log Matching Property**原则就已经保证了选出的leader在整个集群中日志的完整性。

#### applier日志

1. 如果上任leader在提交日志之前宕机，下一任 leader将尝试完成日志的复制。这时候，如果有rf.logs[%v].Term: %v != rf.currentTerm的情况，因为新的leader不能准确判断这个log是不是已经提交，就不能去提交这个log，也就是只提交当前term的日志；
2. 如果在上个term残留的日志后面有新的日志满足提交的条件，因为**Log Matching Property**已经保证了在此之前的日志都是一致性的，那就会把新日志之前的所有log都保证提交，顺便也提交了之前的term残留的log。

### 锁的使用

1. 考虑修改server的状态变量的时候一定要上锁，
2. 在可能需要wait的操作不要上锁：channel的读写，等待timer，sleep()，发送RPC ，重传RPC（已经删除）
3. 先考虑大粒度的锁，不要提前优化，不过大粒度的锁也要考虑死锁问题

Log Matching Property：

1. 如果两个日志的两条日志记录有相同的索引和任期，那么这两条日志记录中的命令必然也是相同的。
2. 如果两个日志的两条日志记录有相同的索引和任期，那么这两个日志中的前继日志记录也是相同的

### 2B-bug记录 

1. append RPC 追加日志成功后重发还是发了日志片段，可能是nextIndex还没改过来就发下一次RPC了：目前推测发的是心跳，也就是所谓的过时RPC
   * 已经解决，心跳发空的日志 
   * 4.26 改成心跳不能空日志， 把 start()的RPC发送取消，交给心跳
2. 旧leader拿到新的log后回来，新leader以为他的nextIndex比实际的小1，而且新leader没有收到新的log所以发的只是心跳。这时候心跳检查, 可以让旧的leaders下台成为follower，log的长度匹配一样，但是因为nextIndex比实际的小1,检查的是前一个，index和term是匹配的，之后准备要把和leader所有不一致的删除，缺失的补上，但是因为是心跳，发过来的entry是空的。
   * 目前解决方案，已经改为心跳和普通同步一致，也就是也包含日志并且把 start() 的提交去掉，等待心跳提交，来减少RPC的超发，因为心跳很密集，同步比较快
3. student guide ：如果选举超时，没有收到现任领导的AppendEntries RPC，也没有向候选人授予投票权：转换为候选人。和我的处理不一样，可能会选举比较慢

### persistence

1. leaderCommit() 的提交log需要检查和当前的term是否一致，如果一致则提交，如果不一致不管。
2. 当leader提交新的log，那么之前的log间接提交，因为 log Matching Property

难改的测试： Figure8C leader commit判断通过之后，还没有applier就断了或者applier了之后断了，这样一千次

1. 因为persist只保存voteFor 、 log 、 currentTerm，所以commitIndex和 lastAppliyIndex在重启的时候还是0，那之前提交过的log现在还需要重头开始重新提交一遍
2. 日志同步的第一次retry没有成功，没有进行下一次retry，很奇怪。应该是死锁，注意锁起来的函数递归调用可能会死锁
   改为 日志同步RPC不 retry，修改nextIndex交给下一次心跳同步

### 快照

用于日志压缩，快照之后，存储当前的状态，那么这个点之前的日志就可以删除。

![image-20230504201733607](D:\MyTxt\typoraPhoto\image-20230504201733607.png)

一般机器单独的进行快照，除非有一个很慢或者新加入的follower需要leader网络发送快照来使其快速追赶

InstallSnapshot RPC

leader发送快照RPC，followers来决定使用；

1. 如果快照包含新信息超过follower的logs, 那会完全选择快照覆盖和logs删减；
2. 如果快照比follower的logs短，那prefix覆盖，之后的保留

**快照和一致性的冲突的**：虽然违背了只有leader修改logs的强领导原则，但是快照的时候一致性已经达成了，所以没有决定是冲突的，数据流还是leader流向follower

快照的时机：确定日志总量达到一个阈值进行快照

### InstallSnapshot RPC

1. 快照RPC发送的时机：

   落后较多的follower在同步日志的catchup过程中，leader回退自己的Log，到起点无法回退，这很好leader发送快照给FollowerFollower，之后通过AppendEntries将后面的Log发给FollowerAppendEntries将后面的Log发给Follower

2. 之前的log没有applier怎么办，直接把快照包装成ApplyMsg，传到applyCh由上层raft处理
3. 日志发送的是到lastIncludedindex位置的前面的全部，没有按照论文所说的实现偏移量

### 压缩后修改index

因为压缩之后logs数组的下标就不是index了，需要大改。。。

1. lastLog的index可能变成了以一个dummy节点，index =0，需要改成nextIncludedIndex，因为在选举的时候需要判断

2. nextIndex 等等也需要改

3. 添加一个接口，然后一个个测，改掉所有的下标

   func (rf *Raft) GetRealLastLogIndex() int {

     return Max(rf.logs[len(rf.logs)-1].Index, rf.LastIncludedIndex)

   }

### 2d-快照bug

1. 先写所有server的快照功能，需要大改下标，主要是加了rf.GetRealLastLogIndex()函数

2. 快照之后，上一次append还没发出去的Entry会被部分覆盖，也就是说在go程去append 的时候args被其他线程修改了

这个bug改了好久，并发的问题，因为之前args.Entry的创建直接用数组切片初始化Entries:    rf.logs[nextindex-rf.LastIncludedIndex:]，可能编译器折叠args.Entry为切片的表达式，也就是并发过程中切片本身变化可能会导致args.Entry的变化？？

之后改成了make空间，然后用copy(args.Entries, rf.logs[nextindex-rf.LastIncludedIndex:])赋值，bug解决

3. crash后恢复的server需要把之前的log重新应用到状态机（这里重新应用logs应该采用提交快照应用），否则直接往后提交新追加的logs是无法提交的，因为lastApplied没有持久化是恢复的时候是0，用日志applier的方式会从log1开始交，但是lastIncludedindex之前的log已经被快照剪短了。

改成make的时候lastApplied和commitIndex都初始化为lastIncludedindex，但是这时候也可能会出现直接提交新的log，而没有实现重新应用之前的log到状态机。这里测试能通过，不清楚什么原因？？？？可能是之前的状态也持久化了？

真正的原因是在2D之后已经**持久化了快照**，那服务器重启的时候**上层状态机就可以直接读取硬盘中存储的快照**（之前的logs的总和操作），对于重启的server只需要继续添加之后的log就可以了。

### 优化

1. 因为持久化的操作在实际过程中是读写硬盘，比较花时间，实验原本给的persist.save()接口只能同时修改state和snapshot，之前在只修改state情况下采用的是保存一遍自己原本的snapshot。
   现在添加persist.saveStateonly()接口，去掉了没有新快照的时候重复读取自己的快照的操作

### 还存在的问题

1. 全部案例一般用时realt time460-470s, CPU时间 2.8s；加上-race 是八分4，CPU时间28s。
2. 没有竞争的全套实验2测试(2A+2B+2C+2D)所需的合理时间是6分钟的实时时间和1分钟的CPU时间。当使用-race运行时，它大约是10分钟的实时时间和2分钟的CPU时间
3. 主要是backup比知道是慢10s，2D不稳定

### git

文件夹有原来关联的仓库，git push不到自己的仓库。

```
git status
//
git init
git remote add origin git@github.com:forthdifferential/6.824.git
git add .
git status

git add .
git commit -m "first commit"
git push
fatal: 当前分支 master 没有对应的上游分支。
为推送当前分支并建立与远程上游的跟踪，使用

    git push --set-upstream origin master

klchen@vmware:~/project/MIT6.824$ git push -u origin main
error: 源引用规格 main 没有匹配
error: 无法推送一些引用到 'git@github.com:forthdifferential/6.824.git'
klchen@vmware:~/project/MIT6.824$ git branch
* master
klchen@vmware:~/project/MIT6.824$ git push --set-upstream origin master

```

这之后上传的文件夹有个箭头，打不开，主要是因为链接了其他的仓库，也就是子文件夹中存在`.git`文件，导致冲突

解决方法： 

```
1. 找到存在的.git文件然后删除，rm .git -r
2. git rm --cached [该文件夹名] git rm --cached 6.5840/
3. 重新执行git add .
4. 执行 git commit -m "rm old_git"
5. 执行git push origin [branch name]  git push
```

### 总结

1. raft集群一般是3 5 个，单数防止脑裂，一个服务器损坏的平均情况大概是几个月一次，所以3 5 个足够修复恢复了
2. 服务器遇到的问题会有网络分区联系不上、机器故障挂了等

有空看看Paxos，因为这是最经典的一致性算法

## Zookeeper

### 线性一致性

关于历史记录的定义，而不是关于系统的定义，也就是说这个系统对外界的请求表现出的是线性一致性。

计算机执行的序列，同时满足一下两个，就说明请求历史数据是线性的

1. 序列中的请求的顺序与实际时间匹配

2. 每个读请求看到的都是序列中前一个写请求写入的值

反之，如果序列的规则生成了一个带环的图，那么请求历史数据不是线性一致性的。

**重点是说**：1. 对于系统执行写请求，只能有一个顺序，所有客户端读到的数据的顺序，必须与系统执行写请求的顺序一致。

2. 对于读请求不允许返回旧的数据，只能返回最新的数据。或者说，对于读请求，线性一致系统只能返回最近一次完成的写请求写入的值。

这点在分布式系统中，不容易保证，因为副本很多，而在单个服务器的情况下天然保证。

* 一个符合线性一致性但是看起来回复不一致的情况：

客户端请求后一直没收到回复，重新发送请求，可能会受到上一次处理的结果，而不是最新的结果。
可能的原因是如果这是因为第一次系统执行完毕了并返回了，而响应故障或者回复报文网络丢包了，系统建立一个表存储为此客户端处理过的结果。

### Zookeeper介绍

Raft实际上就是一个库。你可以在一些更大的多副本系统中使用Raft库。但是Raft不是一个你可以直接交互的独立的服务，你必须要设计你自己的应用程序来与Raft库交互。

Zookeeper看成一个类似于Raft的多副本系统，运行在Zab（和Raft一样）之上。

为了提高系统的性能，需要把除了leader外的服务器用起来，但是线性一致性的要求表示只能通过leader交互，因为不能保证follower的日志是up-to-date。Zookeeper 的做法是不保证线性一致性，把读操作分摊给所有的peer处理，这说明请求一些落后的副本的时候可能会读取到一些旧的数据。

### Zookeeper的一致性保证

1. 保证写请求是线性一致性的，也就是执行是按照一致性的顺序；但是读一致性不考虑，不需要经过leader，可能会返回旧数据
2. 任何一个客户端的请求都会按照顺序执行（通常是看系统不一定会顺序执行的），FIFO，可能的实现是为所有请求打上编号，然后leader节点遵从这个顺序。

FIFO的实现注意：

1. 如果客户端在一个S读取，这个S挂了，切换到下一个S的时候，那从下一个S读取的数据必须要在上一次读取数据的log的后面。 也就是说第二个读请求至少要看到第一个读请求的状态！

​	  如果在这个位置的log之前，那新的S会阻塞客户端的响应或者告诉他找其他的S试试。

2. 对于单个的客户端，因为FIFO的要求，所以读写的线性一致性实际上能保证。比如一个写一个读，写给leader 17，读给了follower读，要阻塞到写完17才能读到17 ，不然客户端收到的回复很奇怪。

### Zookeeper的同步操作 （sync）

弥补非严格线性一致的方法是设计了一个操作类型是sync，效果相当于写请求。

如果我需要读最新的数据，我需要发送一个sync请求，之后再发送读请求。这个读请求可以保证看到sync对应的状态，所以可以合理的认为是最新的。但是同时也要认识到，这是一个代价很高的操作，因为我们现在将一个廉价的读操作转换成了一个耗费Leader时间的sync操作

### Zookeeper应用场景

如果你有一个大的数据中心，并且在数据中心内运行各种东西，比如说Web服务器，存储系统，MapReduce等等。你或许会想要再运行一个包含了5个或者7个副本的Zookeeper集群，因为它可以用在很多场景下。之后，你可以部署各种各样的服务，并且在设计中，让这些服务存储一些关键的状态到你的全局的Zookeeper集群中。

### Zookeeper中watch通知

大体上讲 ZooKeeper 实现的方式是通过客服端和服务端分别创建有观察者的信息列表。客户端调用 getData、exist 等接口时，首先将对应的 Watch 事件放到本地的 ZKWatchManager 中进行管理。服务端在接收到客户端的请求后根据请求类型判断是否含有 Watch 事件，并将对应事件放到 WatchManager 中进行管理。

在事件触发的时候服务端通过节点的路径信息查询相应的 Watch 事件通知给客户端，客户端在接收到通知后，首先查询本地的 ZKWatchManager 获得对应的 Watch 信息处理回调操作。这种设计不但实现了一个分布式环境下的观察者模式，而且通过将客户端和服务端各自处理 Watch 事件所需要的额外信息分别保存在两端，减少彼此通信的内容。大大提升了服务的处理性能。

需要注意的是客户端的 Watcher 机制是一次性的，触发后就会被删除。

### API

Znode类型

- regular：常规节点，它的容错复制了所有的东西
- ephemeral：临时节点，节点会自动消失。比如session消失或Znode有段时间没有传递heartbeat，则Zookeeper认为这个Znode到期，随后自动删除这个Znode节点
- sequential：顺序节点，它的名字和它的version有关，是在特定的znode下创建的，这些子节点在名字中带有序列号，且节点们按照序列号排序（序号递增）。

一些API用RPC调用暴露

### Ready file(znode)

Master节点在Zookeeper中维护了一个配置，这个配置对应了一些file（也就是znode）。通过这个配置，描述了有关分布式系统的一些信息，例如所有worker的IP地址，或者当前谁是Master。所以，现在Master在更新这个配置，同时，或许有大量的客户端需要读取相应的配置，并且需要发现配置的每一次变化 。

尽管配置被分割成了多个file，还需要保证有原子效果的更新：

如果Ready file存在，那么允许读这个配置。如果Ready file不存在，那么说明配置正在更新过程中，我们不应该读取配置。所以，如果Master要更新配置，那么第一件事情是删除Ready file。之后它会更新各个保存了配置的Zookeeper file（也就是znode），这里或许有很多的file。当所有组成配置的file都更新完成之后，Master会再次创建Ready file。

### Zookeeper实现计数器

处理并发的客户端请求，不能用单纯的读写数据库的方式，而且Zookeeper本身可能返回旧值，这也要考虑。

```
    WHILE TRUE:
    X, V = GETDATA("F")
    IF SETDATA("f", X + 1, V): // leader节点执行 ，要求读到最新的数据且版本号还没被修改
        BREAK
```

方式是min-transaction，mini版的事务的实现，把读-更改-写作为一个原子操作

## lab3

FT-kv存储，是复制状态机；

clien向k/v service发送三个RPC：Put Append Get，其中是Clerk 与 Client交互，Clerk通过RPC与servers交互。

Get/Put/Append 方法需要是线性的，对外表现是一致的，我理解就是说并发情况下 对外表现的线性一致性：

后一次请求必须看到前一次的执行后端状态；并发请求选择相同的执行顺序，避免不是最新状态回复客户端；在故障之后保留所有确认的客户端更新的方式恢复状态

### 3A

每个kv service有一个raft peer，Clerks把Get/Put/Append  RPC发送给leader的kv service，进一步交给Raft，Raft日志保存这些操作，所有的kv service按顺序执行这些操作，应用到kv数据库，达到一致性

1. Clerk找leader所在kv service和重试RPC的过程；
2. 应用到状态机后leader通过响应RPC告知Clerk结果，如果操作失败（比如leader更换），报告错误，让他重试
3. kv service之间不能通信，只有raft peer之间RPC交互

### 构思：

主要过程： Clerk把操作包装Op发送 PRC给 所有server，如果server是leader就开始发给raft，raft同步完成后apply给server，server执行这个Op,返回给Clerk结果。

如果raft的applier超时，RPC回复超时，因为是旧leader，所以等到下一个term再发送RPC给新leader的server。

### 重复RPC检测

如果收不到RPC回复（no reply）,一种可能是server挂了，可能换一个重新请求；但是另一种是执行了但是 reply丢包了，这时候重新发的话会破坏线性一致性。 

解决方案是重复RPC检测：

clerk每次发RPC都发一个ID，一个Request一个ID，重发相同；在server中维护一个表记录ID对应的结果，提前检测是否处理过；Raft的日志中也要存这个ID，以便新的leader的表是正确的

如果之前的请求还没执行，那会重新start一个，然后等第一个执行完后表就有了，applCh得到第二个的时候看表再决定不执行了。也就是有两个一样的log 没关系的

### 请求表的设计

每个客户端一个条目，存着最后一次执行的RPC：

每个client同时只有一个未完成的RPC，每个client对PRC进行编号；也就是说当客户端发送第10的条目，那之前的都可以不要，因为之前的RPC都不会重发了；

执行完成后才更新条目，index是clientID，值是保留值和编号，

### client

每个client有一个clientID，是一个64位的随机值 

client发送PRC中有 clientID 和 rpcID ，重发的RPC序号相同



如果执行完被leaderID索引，内容仅包含序号和值

RPC处理程序首先检查表格，只有在序号>表格条目时才Start()s

每个日志条目必须包括客户端ID、序号

当操作出现在applyCH上时，更新client's table entry中的序号和值，唤醒正在等待的RPC处理程序(如果有)



### 3A-bug

1. 一个clerk就是一个client调用的，所以clientId应该在clerk结构里，之前我是 每个请求随机生成了一个clientId，这是不太合理的

2. 处理RPC超时不是 !ok表示的，应该要自己设置时间。本来使用的是条件变量等着表的更新，但是用select chanl可以用非阻塞处理处理不同的通道，而且可以直接传递变量，不需要重新查表了。 原代码

   ```cpp
   	if !isLeader {
   		reply.Success, reply.Err = false, "not leader"
   		return
   	}
   
   	for !kv.killed() {
   		if equset_entry, ok := kv.requsetTab[args.ClientID]; ok {
   			if args.RpcID <= equset_entry.rpcID {
   				// 已经执行RPC请求
   				reply.Success, reply.Value = true, equset_entry.value
   				return
   			}
   		} else {
   			DPrintf("SERVER {server %v}等待一致性通过,Optype: Get,ClientID: %v,RpcID: %v,Key %v", kv.me, args.ClientID, args.RpcID, args.Key)
   			kv.receiveVCond.Wait()
   		}
   	}
   ```

   定时器控制RPC超时 ，chanl传入处理完的数据

   疑问：定时器应该是一个函数一个还是放到server结构里，chanl也是一个函数一个，那开销好像比较大

   用 map[int]chan int处理 对应不同的index，提交时到对应chan中，index的唯一性。

   考虑添加chan和删除，delete(kv.chanMap, 1)   删除键为1的元素，自动销毁chan，考虑请求满足之后就会删除，同时存在chan也不多，先采用这种方案

2. 设计是定时器在server中，那server如果挂了，也是没有reply，不返回了，这种情况怎么解决？

这是服务器的问题，不是raft一致性没有达到，应该是RPC会返回false。

4. 看起来client认为的server编号和server自己编号不一致，把前者不打出log以免出错

   问题还有是重复检测发生在检查leader之前了，那可能不是leader但因为同步执行到了，也能返回，然后client更新了错误的leader。

   修改了原来的结构，改为先检查leader，然后重复检测，然后再start

5. **难改bug**TestSpeed3A 要求每个心跳周期至少处理三次client的请求，但是我的raft中，start收到新条目不会马上同步，而是等下一次心跳来同步。这个TestSpeed3A 是同一个客户端发送一千次append，RPC接受函数阻塞在那直到apply后回复，而我每次都是一个心跳处理一次，就会导致一个心跳周期值处理一个请求。

​		具体表现是：rpc一致性通过，执行完成，返回通道的时候超时了，然后一直重复超时

先改为start( )之后就发起心跳，试试

server处理操作的PRC超时时间设置有点不清楚，改到800ms这样大可以减小TestSpeed3A 不通过的概率，但是跑多了还是会有个错误。只可能原因是高并发多请求的时候apply处理不过来了

5.2 这中间还有一个bug，就是raft一致性达成了而且apply了太快了，通道还没make。那就不传这个通道，等超时返回，并在下一次RPC的重复检验的时候直接取到值

6. applCh拿到后发现表中已经有rpc，没有执行也没有返回，导致RPC得不到回复。

正确的做法是不执行，但是要返回值。查看rpc通道还存不存在，如果存在，说明上一次没有回复rpc（可能是上次chanl还没创建好）；如果不存在，说明这个条目是重复的条目，不需要返回RPC了

7. 好看点的写法是函数中间需要加锁的部分单独移出去做新函数，不然很丑而且容易忘记解锁
8. **严重bug** - leader的applyCh读取不到数据了，导致一直超时，一直操作不了数据库。 

原来是raft日志同步有问题，也就是在start后立即发起心跳引起的。

也就是修改后导致添加了额外的心跳，那在发送心跳的过程中snapshot了，会导致preLogIndex比follower的lastIncludeIndex还要小，这种情况我提前考虑，然后返回xindex为realLastlogIndex

9. ***改好久bug***leader 提交raft后没有达成了一致性但是没有apply上来，反倒是其余follower都apply上来了。

检查了一圈，问题应该是raft在同步完成后，leader检查可否更改commitIndex，如果通过更改commintIndex，那会用条件变量唤醒apply，但是实际上到这一步后没有成功唤醒apply。

原因是rf.applyVCond.Broadcast()唤醒不了在另一个go程上的applier，这种情况发送的不多，但是每次发送都是永远提交不上，原因也不知道。

我才有一种可能是Broadcast()的时候，那个线程不在wait。所以把commitLeader()中唤醒的操作放到从之前的 每次增加commit唤醒 改到 函数最后再唤醒，这样减少applier压力。但是这样TestSpeed有时候过不了，还是改回原来，考虑其他优化

考虑没有进入wait状态就阻塞的原因，去找阻塞再哪一句，发现在写入chan的时候阻塞了。有可能是server一直在读导致的异常阻塞，把server改成select非阻塞读取。

另外chan写端阻塞可能是缓冲区的为0是非缓冲的，我把applyCh和map中的chan改为缓冲区是1。估计这个才是主要原因。



### 3B

### 要求： 

测试程序将MaxraftState传递给您的StartKVServer()。

MaxraftState指示持久RAFT状态允许的最大大小(以字节为单位)(包括日志，但不包括快照)。**将MaxraftState与Persister.RaftStateSize()**进行比较。当您的键/值服务器检测到RAFT状态大小接近此阈值时，它应该通过调用RAFT的Snapshot来保存快照。

如果MaxraftState为-1，则不必创建快照。

MaxraftState应用于RAFT作为第一个参数传递给Persister.Save()的GOB编码的字节。

server重启的时候，从persister读取快照并且恢复

1. 快照需要存储的信息和快照的时机考虑清楚，kvserver的重复检验必须跨越检查点，所以任何状态必须包含在快照中。可能快照需要请求表因为一恢复就要启动重复检验的功能，应该是需要保存kv数据库，也就是状态机的状态
2. 快照中数据大写
3. raft可能被改错，记得常回看看

### 构思：

#### 快照时机：

考虑server和raft交互的时候，考虑server状态机和raft日志改变的时候。

在raft通过applyCh提交给server的时候，检查是否快照。快照的时候设置这个index之前，所以在应用状态机之前检查快照

#### 快照操作：

1. 检查stateSize，如果超出就做快照。
2. server做快照，需要保存数据库和请求表snapshot，保存index（因为之后还要传给让raft，且applyCh可能接受快照的，那时候需要比较快照的latest性）。然后把snapshot和index传给raft做快照，也就是保存raftstate
3. 恢复操作，server恢复状态机、请求表，raft恢复state
4. appCh可能会传条目，也可能传快照，分开处理。如果传的是快照，之前raftstate已经更新，判断快照合理性，应用快照。

### 3B-bug

1. raft提交一个snapshot并被server应用之后，又提交了一个相同index的Op，但内容是nil

原因的是server主动快照时在新applyCh传入条目的应用之前拍的，也就是传入快照函数的index是新条目的index-1而不是index 

2. 落后较多raft接受到RPC快照同步的时候，没有成功。仔细看原因是applCh写入snapshot的时候阻塞了。
2. raft下标越界，可能原因是用chan的时候解锁，锁被抢走了，修改了lastIncludeIndex,当chan写完获取锁的时候，已经改变了raft的状态。

对2 3bug的修改，主要是把安装快照的RPC中传给上层的chan放到最后处理，这时候PRC导致的状态更新已经完成了，chan放锁给其他的用来修改raft状态也没关系。

所以说chan不仅要记得解锁，还锁，而且要考虑在chan解锁期间发送的不确定性问题

