# schat
A Simple Chat Serv
一个简单的聊天系统服务端

### 说明  
基于互联网的应用，无论用户如何变化与分散，所有的信息与交流都会汇集于服务器，这在无形之中也不可避免地造成了服务端的利维坦化。即服务提供方将接管所有用户的兴趣，需求与内容。这可能会导致普通人的信息污染与泄漏，同时导致不安全的事情发生。由于个人理想不小但能力不大，所以设计这一款简单的服务框架一个小的目的，就是将服务器的部署与运行也由用户来自己控制(当然可能需要一台云服务器)，从而尽量减少在一些隐私信息交流过程中的泄漏。目前这是服务端框架，尽力提供了简单部署，简单迁移，简单恢复和简单爱，谋求用最小的成本来实现简单化的群聊系统

### Features
* **简单安装** 基于SGAME框架开发生成，在完成了SGAME的环境搭建之后一键安装。sgame的开发环境搭建请参考 https://github.com/nmsoccer/sgame 的安装说明  
* **简单部署** 作为示例全部业务进程在本机部署，修改配置文件可以方便的部署到不同机器上  
* **简单迁移** 在一个服务器系统失效之后可以方便的新开服务器并迁移数据  
* **简单恢复** 迁移之后仍可维持原有的聊天数据及成员信息  


### 架构
由不同的业务SET构成，SET内的进程均可平行扩展  
![架构](https://github.com/nmsoccer/schat/blob/master/pic/schat.jpg)



