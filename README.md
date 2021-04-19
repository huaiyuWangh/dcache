# dcache

分布式缓存系统
    基于lru的缓存淘汰策略
    基于consistenthash算法实现节点选择
    基于consul的服务发现实现节点的动态变更
    通过singleflight解决缓存击穿问题