# schat
A Simple Chat Serv
一个简单的聊天系统服务端

### 说明  
基于互联网的应用，无论用户如何变化与位置分散，所有的信息与交流都会汇集于服务器，这在无形之中也不可避免地造成了服务端的利维坦化。即服务提供方将接管所有用户的兴趣，需求与内容。这可能会导致普通人的信息污染与泄漏，同时导致不安全的事情发生。引用一下弗雷泽的观点，正常的公民生活需要“自我—管理(self－management)、交互—公共领域合作(inter－public coordination)以及政治责任(political accountability)” 。设计这一款简单的服务框架的初衷，就是将服务器的部署与运行尽量解放出来由用户来自己控制(当然可能需要一台云服务器)，从而尽量减少在一些在次群(sub group)交流过程中导致的隐私信息泄漏.目前这是服务端框架，尽力提供了简单部署，简单迁移，简单恢复，谋求用最小的成本来实现简单化的群聊系统

### Features
* **简单安装** 尽量减少外部依赖库的使用  
* **简单部署** 作为示例全部业务进程在本机部署，修改配置文件可以方便的部署到不同机器上  
* **简单迁移** 在一个服务器系统失效之后可以方便的新开服务器并迁移数据  
* **简单恢复** 迁移之后仍可维持原有的聊天数据及成员信息  

### 功能
整个聊天系统以群聊为核心，所有功能均围绕群展开，目前包括以下功能：
* **账号注册** 以名字为Key申请注册账号
* **创建群组** 注册成功的用户可以创建聊天群和入群密码
* **申请入群** 申请加入群
* **审批申请** 群主审批请求，同意或拒绝
* **群内聊天** 入群之后可以发送文本，图片等
* **退出群聊** 群员可以主动退群
* **踢出群聊** 群主可以踢出成员
* **解散群组** 群主可以解散群


### 架构
由不同的业务SET构成，SET内的进程均可平行扩展  
![架构](https://github.com/nmsoccer/schat/blob/master/pic/schat.jpg)


### 进程
对拓扑图里的进程功能及部署进行说明
* **conn_serv**   
  客户端接入进程，负责维护客户端的连接，客户端使用TCP长链接接入  
* **logic_serv**  
  在线用户的数据缓存，处理用户本身的主要逻辑服务  
* **db_logic_serv**  
  与logic_serv配对的db代理进程，负责与reddis的连接与数据交互
* **说明1**  
  conn_serv,logic_serv,db_logic_serv一般1:1:1配置对应作为一个用户连接，处理和数据的逻辑单元，按逻辑单元平行扩展  
* **disp_serv**    
  作为星形拓扑的包分发中心，负责分派各业务进程之间的数据包转发，这样每个业务进程不需要维护其他多余的进程通信地址，只需要和disp进程组连接即可。一般需要与其他业务进程组互相通信的进程组与disp_serv进程组建议通信；disp_serv可以平行扩展  
* **online_serv**  
  缓存世界里所有在线用户的logic_serv地址，一般部署两个serv即可，双主作为互备  
* **file_serv**   
  静态文件服务进程，目前主要有两种功能：1.将群聊内发布的文件包括图片等存储，同时生成对应的聊天记录；2.存储用户头像。进程可以设置安全等级，作为一般的服务验证。每个file_serv需要配置唯一的servindex作为文件url的一部分，同时方便数据迁移而保持所有群聊文件数据。文件服务进程亦可平行扩展，更具体说明可以参考wiki  
* **chat_serv**  
  聊天管理进程，这里会缓存所有活跃(主要是聊天等)的群组数据，群组数据按群ID hash分布到chat_serv上。同时用于同步转发聊天信息.
* **db_chat_serv**  
  服务于chat_serv的db代理，一般与chat_serv 1:1配置作为一个逻辑处理单元，平行扩展时最好按处理单元扩展  
* **dir_serv**  
  用于connect_serv前端的负载均衡，同时作为file_serv的相关地址信息管理  



### 环境安装
schat基于sgame框架，所以其安装环境与sgame流程一致，这里摘自https://github.com/nmsoccer/sgame：  
#### 基础软件
* **GO**  
下载页面https://golang.google.cn/dl/ 或者 https://golang.org/dl/  这里下载并使用go 1.14版本，然后
  * tar -C /usr/local -xzf go1.14.6.linux-amd64.tar.gz  
  * 修改本地.bashrc export PATH=$PATH:/usr/local/go/bin export GOPATH=/home/nmsoccer/go 

* **PROTOBUF**  
下载页面https://github.com/protocolbuffers/protobuf/releases  这里选择下载protobuf-all-3.11.4.tar.gz.
  * 解压到本地后./configure --prefix=/usr/local/protobuf; make; make install  
  * 修改本地.bashrc export PATH=$PATH:/usr/local/protobuf/bin

* **REDIS**  
下载页面https://redis.io/download 这里选择下载redis-5.0.8.tar.gz. 
  * 解压到本地后make 然后拷贝src/redis-cli src/redis-server src/redis.conf 到/usr/local/bin.
  * 修改/usr/local/bin/redis.conf新增密码requirepass cbuju 用作sgame使用redis的连接密码 
  * 修改port 6698作为监听端口 然后cd /usr/local/bin; ./redis-server ./redis.conf & 拉起即可  

#### 必需库
* **PROTOBUF-GO**  
probotuf-go是protobuf对go的支持工具，这里用手动安装来说明.
  * 下载安装 进入https://github.com/protocolbuffers/protobuf-go 下载protobuf-go-master.zip 
    * mkdir -p $GOPATH/src/google.golang.org/; cp protobuf-go-master.zip $GOPATH/src/google.golang.org/; 
    * cd $GOPATH/src/google.golang.org/; 解压并改名解压后的目录为protobuf: unzip protobuf-go-master.zip; 
    * mv protobuf-go-master/ protobuf/
  * 生成protoc-gen-go工具 进入$GOPATH/src 
    * go build google.golang.org/protobuf/cmd/protoc-gen-go/ 顺利的话会生成protoc-gen-go二进制文件   
    * mv protoc-gen-go /usr/local/bin   
  * 安装proto库 进入$GOPATH/src 
    * go install google.golang.org/protobuf/proto/ 安装协议解析库
  * 完成 进入任意目录执行protoc --go_out=. xxx.proto即可在本目录生成xxx.pb.go文件（这里使用proto3）
  
  * 补充安装github.com/golang/protobuf/proto 由于生成的xx.pb.go文件总会引用github.com/golang/protobuf/proto 库，所以我们一般还需要额外安装这个玩意儿. 
    * 进入https://github.com/golang/protobuf 页面，下载protobuf-master.zip到本地.
    * mkdir -p $GOPATH/src/github.com/golang/目录. cp protobuf-master.zip $GOPATH/src/github.com/golang/. 
    * 解压并重命名:cd $GOPATH/src/github.com/golang/; unzip protobuf-master.zip; mv protobuf-master/ protobuf/; 
    * 安装 cd $GOPATH/src; go install github.com/golang/protobuf/proto
  
* **REDIGO**    
redigo是go封装访问redis的支持库，这里仍然以手动安装说明
  * 下载  
  进入https://github.com/gomodule/redigo 页面,下载redigo-master.zip到本地
  * 安装  
    * 创建目录mkdir -p $GOPATH/src/github.com/gomodule; cp redigo-master.zip $GOPATH/src/github.com/gomodule   
    * 解压并重命名: cd $GOPATH/src/github.com/gomodule; unzip redigo-master.zip; mv redigo-master redigo  
    * 安装: cd $GOPATH/src; go install github.com/gomodule/redigo/redis  
  
* **SXX库**  
sxx库是几个支持库，安装简单且基本无依赖,下面均以手动安装为例  
  * slog  
  一个简单的日志库.https://github.com/nmsoccer/slog. 下载slog-master.zip到本地，
    * 解压然后安装:cd slog-master; ./install.sh(需要root权限)
  * stlv
  一个简单的STLV格式打包库. https://github.com/nmsoccer/stlv. 下载stlv-master.zip到本地   
    * 解压然后安装:cd stlv-master; ./install.sh(root权限)
  * proc_bridge
  一个进程通信组件，sgame里集成了proc_bridge，这里需要安装支持库即可. https://github.com/nmsoccer/proc_bridge 下载proc_bridge-master.zip到本地  
    * 解压安装:cd proc_bridge-master/src/library; ./install_lib.sh(root权限)，安装完毕. 更加详细的各种配置请参考https://github.com/nmsoccer/proc_bridge/wiki

### SCHAT安装  
这里仍然以手动安装为例
  * 下载安装    
  进入 https://github.com/nmsoccer/schat; 下载schat-master.zip到本地; 
    * 部署cp schat-master.zip $GOPATH/src/; cd $GOPATH/src; unzip schat-master.zip; mv schat-master schat 完成  
  * 配置通信  
    * 进入 $GOPATH/src/schat/proc_bridge. (这里的proc_bridge就是上面安装的proc_bridge组件，只是为了方便集成到这个项目里了).然后执行./init.sh初始化一些配置.
    * 进入schat/目录。 修改bridge.cfg配置，因为我们是本机部署，所以只需要修改BRIDGE_USER，BRIDGE_DIR这两个选项使得用户为本机有效用户即可.具体配置项请参考https://github.com/nmsoccer/proc_bridge/wiki/config-detail
    * 执行 chmod u+x build.sh; ./build.sh install  
    * 执行 ./manager -i 1 -N schat 这是一个通信管理工具 执行命令STAT * 可以查看到当前路由的建立情况. 执行PROC-CONN * 可以查看是否有网络连接错误。 具体使用可以参考https://github.com/nmsoccer/proc_bridge/wiki/manager  

  * 编译进程
    * 进入$GOPATH/src/schat/servers/spush chmod u+x init.sh build_server.sh
    * 执行./init.sh 初始化设置
    * 执行 ./build_servers.sh -b 编译(也可以进入servers/xx_serv各目录下go build xx_serv.go 手动编译)

  * 发布进程
    * 进入$GOPATH/src/schat/servers/spush
    * spush是一个分发管理工具，下载自https://github.com/nmsoccer/spush 这里也将其集成到了框架内部
    * schat.json 是spush使用的配置文件，我们都是本地部署所以只需要schat.json文件里的nmsoccer用户名配置成本机有效用户xxx即可
      sed -i "s/nmsoccer/xxx/g" schat.json
    * 发布拉起 
      ./spush -P -f schat.json 结果如下:
      ```
      ++++++++++++++++++++spush (2020-10-02 19:54:03)++++++++++++++++++++
      push all procs
      .create cfg:18/18
      ................
      ----------Push <schat> Result---------- 
      ok
      .
      [18/18]
      [manage_serv-1]::success 
      [logic_serv-1]::success 
      [db_logic_serv-1]::success 
      [chat_serv-1]::success 
      [disp_serv-2]::success 
      [db_chat_serv-2]::success 
      [online_serv-1]::success 
      [online_serv-2]::success 
      [dir_serv-1]::success 
      [conn_serv-2]::success 
      [chat_serv-2]::success 
      [db_chat_serv-1]::success 
      [disp_serv-1]::success 
      [file_serv-1]::success 
      [file_serv-2]::success 
      [conn_serv-1]::success 
      [logic_serv-2]::success 
      [db_logic_serv-2]::success 

      +++++++++++++++++++++end (2020-10-02 19:54:20)+++++++++++++++++++++
      ```
      说明OK鸟  
    
    * 关闭全部进程    
      一般进入页面端进行关闭
    
    * 单独推送全部进程配置  
      ./spush -P -f schat.json -O  
    
    * 单独推送全部进程BIN文件  
      ./spush -P -f schat.json -o
      
    * 单独推送某个进程BIN文件及配置      
      ./spush -p ^logic* -f schat.json    
      更多spush的使用请参考https://github.com/nmsoccer/spush   
  
    * 页面监控
      * 如果拉起进程顺利，我们可以打开页面查看，默认端口是8080同时需要用户名及密码,默认选项配置于spush/tmpl/manage_serv.tmpl:auth,登陆查看：
    ![管理页面](https://github.com/nmsoccer/schat/blob/master/pic/schat_index.png)   
  
* **未完待续**  
