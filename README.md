 # cmdb_agent
 cmdb  客户端Agent  采集信息agent

 #cmdb_agent
 Golang版本：go1.16.3

  Agent功能：
  1、轻量采集主机基础信息：cpu 、内存 、disk 、ip...
 
  2、支持扩展自定义命令返回执行命令结果信息
 
  3、支持配置http 推送采集结果
 
  4、支持远程控制Agent start、restart、stop
 
  5、支持远程控制生成Agent配置文件
 
  6、支持邮件告警，以及扩展你的IM、短信告警渠道


打包、启动：
go build
go cmdb_agent -start -config="config.ymal"
测试：http://0.0.0.0:8999/one


几种采集架构设计方案和选择：
***************************
一、Agent （PUSH）


                  cmdb采集模块----------cmdb资源模块
                        |
            cron------metric----resource
                        |
          ip1（http）------ip2(http)
            ||
       config------上报json（时间戳）
        |
    heartbeat --采集项 --- cron

设计说明：
1、定时任务： 自定义、算法生成时间
2、部署：GO、python基础版本部署；一次部署；
3、单机采集metric自定义：Yes
4、配置文件：需要根据APP metric更新；git 维护或者 http推送
5、上报json: 保存上一次采集结果，和下一次采集结果比对;
判断是否上报； 时间戳命名，判断下一次心跳时间。
6、单机采集控制：Yes
*****************************



*****************************
二、Salt  (PULL)

                    (资源比对)
      cmdb采集模块-----------------cmdb资源模块
        |
  cron------metric
        || (http)
    ServerAPI
        || (http)
    ip1-----ip2

设计说明：
1、定时任务： 自定义、算法生成时间
2、部署：多系统、多版本、多操作系统部署；一次部署；
3、单机采集metric自定义：Yes
******************************

三、open-falcon plugin （PULL）


                                                
                  falcon     cron -------- cmdb采集模块 -------- cmdb资源模块 
                    |         |(PULL)
          plugin--------falcon-api


设计说明：
前提：cmdb主机都在falcon监控范围内。
1、定时任务： falcon-agent、plugin 定时上报
2、部署：git 维护plugin 、falcon hostgroup挂载plugin 
3、单机采集metric自定义：No
4、插件管理：angent 配置文件支持一个Git URL。采集脚本目录 open-falcon 共享
5、[{"endpoint": "node1", "tags": "", "timestamp": int(time.time()), "metric": "agent.cpu", "value": 1.8, "counterType": "GAUGE", "step": 60, "note":"json采集数据"}]
6、falcon API: /alarm/eventcases
***********************************



表结构设计
***************************************
采集主表  
资源ip   cpu    memory    disk   状态
******************************************
采集副表：采集命令 管理
资源ip   metric/cmd   采集值  采集结果处理插件  处理结果  状态   最后采集时间
*********************************************
机器采集结果变更日志
资源ip   采集结果   时间
****************************************






cron 管理
资源ip  cron设置时间  状态
********************************************



