>**stress-testing-gui**

> 压测GUI，前端模拟了postman发起请求，渣渣后端写前端，bootstrap一把梭，可自行修改。  
> 填写正确的接口地址，以及需要的并发或者QPS，请求由后端程序处理。
> 并发模式会在后端构建请求的CURL请求体；QPS模式后端程序会在压测时间内向目标接口发送指定数量请求QPS约等于总请求数/请求时间。
 
>* 第一版目前只支持HTTP请求,请求方式目前开放了GET, POST
>* 命令行版本☞☞☞ https://github.com/applytogetintoyourlife/stress-testing（支持变参压测）
>* 作者邮箱 code301@163.com 欢迎交流
>* 部分功能截图
![image](https://user-images.githubusercontent.com/42194819/128487170-54c5e765-6e52-435a-8394-71c0a659e0e5.png)
![image](https://user-images.githubusercontent.com/42194819/128487463-7255829c-e4b3-44ba-80b6-a87efa589acf.png)
![image](https://user-images.githubusercontent.com/42194819/129359898-02c207e7-0600-4d81-b1c5-192f7410eee6.png)
![image](https://user-images.githubusercontent.com/42194819/129360057-4dbdbc5d-4196-413f-b519-f6348d44038e.png)
