# 简介
这是一个示例项目, 使用sqlite作为数据库

```text
# 目标最大承载100万用户的架构
# 数据库存储容量：假设sqlite文件大小为1T，每个用户1M数据，则总的用户量为1024*1024 = 1048576
# 1M约等于512篇，1024字的文章
# 以二八定律计算，有20%活跃用户，每个用户100次查询，则PV= 200000 * 100 / 0.8 = 2500000.0。20%活跃时间，则 QPS ~= 115.74
# 以i5-4460@3.2G，普通机械硬盘， 建索引的sqlite为例
# sqlite 读性能：
# sqlite 写性能：62471 （插入）
```

```text
选择sqlite的理由：
Is the data separated from the application by a network? → choose client/server
Many concurrent writers? → choose client/server
Big data? → choose client/server
Otherwise → choose SQLite!
```

> sqlite性能测试，与其它嵌入型数据库的对比
>　https://blog.csdn.net/tjcwt2011/article/details/110005320

```text
众所周知SQLlite写速度慢，但是有一些优化方法：
https://blog.csdn.net/Ango_/article/details/122074816
1. 关闭写同步
2. 使用事务批量写
3. 执行准备
4. 内存模式

其它优化方法：
使用分库
```

# 环境
操作系统：Windows 11 x64

语言环境：go 1.21+

工具链：TDM GCC:　https://jmeubank.github.io/tdm-gcc/



