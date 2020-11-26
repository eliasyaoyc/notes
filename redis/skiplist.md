# SkipList (跳跃表)
有序集合比如对学生进行排名、根据得分对游戏玩家进行排名等，对于有序集合的底层实现，可以使用数组、链表、平衡树等结构。
* `数组`：不便于元素的插入和删除
* `链表`：查询效率低，需要遍历所有元素
* `平衡树或者红黑树`：效率高但是实现复杂
* `跳跃表`：效率堪比红黑树，实现比红黑树简单，其查询、插入、删除的平均复杂度都为O(logN)，主要应用于有序集合的底层实现。
   * 不使用红黑树的原因除了复杂还有在Redis 中 zset 支持范围查询，红黑树在做范围查询的时候性能太低。
    
跳跃表的实现过程：请看 doc 中的 《为什么Redis一定要用跳表来实现有序集合？》

## Redis 中的实现 
跳表节点 zskiplistNode 结构体
```c
typedef struct zskiplistNode{
   sds ele;
   double score;
   struct zskiplistNode *backward;
   struct zskiplistLevel {
        struct zskiplistNode *forward;
        unsigned int span;  
   } level[];
} zskiplistNode
```
* ele: 用于存储字符串类型的数据
* score：用于存储排序的分值
* backward：后退指针：只能指向当前节点最底层的前一个节点，头节点和第一个节点-backward指向null，从后向前遍历跳跃表时使用
* leve：柔性数组，每个节点的数组长度不一样，在生成跳跃表节点时，随机生成一个1~64的值，值越大出现的概率越低
  * forward：指向本层下一个节点，尾节点的forward 指向null
  * span: forward指向的节点与本节点之间的元素个数。span值越大，跳过的节点个数越多。
  
>跳跃表是Redis有序集合的底层实现方式之一，所以每个节点的ele存储有序集合的成员member值，score存储成员score值。所有节点的分值是按从小到大的方式排序的，当有序集合的成员分值相同时，节点会按member的字典序进行排序。

跳跃表 zskiplist 结构：
```c
typedef struct zskiplist{
  struct zskiplistNode *header,*tail;
  unsigned long length;
  int level;
} zskiplist
```
* header: 指向跳跃表头节点。头节点是跳跃表的一个特殊节点，它的level数组元素个数为64。头节点在有序集合中不存储任何member和score值，ele值为NULL, score值为0；也不计入跳跃表的总长度。头节点在初始化时，64个元素的forward都指向NULL, span值都为0。
* tail: 指向跳跃表尾节点
* length: 跳跃表长度，表示除头节点之外的节点总数
* level: 跳跃表的高度

## Redis 中有关于zset的配置：
* `zset-max-ziplist-entries 128`：zset采用压缩列表时，元素个数最大值。默认值为128
* `zset-max-ziplist-value 64`：zset采用压缩列表时，每个元素的字符串长度最大值，默认为64
```c
/* This generic command implements both ZADD and ZINCRBY. */
void zaddGenericCommand(client *c, int flags) {
    // ... 省略
    /* Lookup the key and create the sorted set if does not exist. */
    zobj = lookupKeyWrite(c->db,key);
    if (zobj == NULL) {
        if (xx) goto reply_to_client; /* No key + XX option: nothing to do. */
        if (server.zset_max_ziplist_entries == 0 ||
            server.zset_max_ziplist_value < sdslen(c->argv[scoreidx+1]->ptr))
        {
            zobj = createZsetObject(); // 创建跳跃表结构
        } else {
            zobj = createZsetZiplistObject(); // 创建压缩列表结构
        }
        dbAdd(c->db,key,zobj);
    } else {
        if (zobj->type != OBJ_ZSET) {
            addReply(c,shared.wrongtypeerr);
            goto cleanup;
        }
    }

    for (j = 0; j < elements; j++) {
        double newscore;
        score = scores[j];
        int retflags = flags;

        ele = c->argv[scoreidx+1+j*2]->ptr;
        // 添加元素
        int retval = zsetAdd(zobj, score, ele, &retflags, &newscore);
        if (retval == 0) {
            addReplyError(c,nanerr);
            goto cleanup;
        }
        if (retflags & ZADD_ADDED) added++;
        if (retflags & ZADD_UPDATED) updated++;
        if (!(retflags & ZADD_NOP)) processed++;
        score = newscore;
    }
   // 省略
```
* 如果把 `zset_max_ziplist_entries` 的值设置为0 的话，那就直接创建跳跃表，但是默认时压缩列表。
* 调用 `zsetAdd` 方法添加元素，在 `zsetAdd` 方法中 会进行条件判断
  * zset 中元素是否大于 `zset_max_ziplist_entries`
  * 插入元素的字符串长度是否大于 `zset_max_ziplist_value`
  * `zsetConvert`：当满足任意条件时，Redis 便会将底层zset的实现由压缩链表转为跳跃表。（Note：转换成跳跃表之后，即使元素被删除，也不会重新转为压缩链表）。
```c
 // zsetAdd
 if (zzlLength(zobj->ptr) > server.zset_max_ziplist_entries ||
                sdslen(ele) > server.zset_max_ziplist_value)
                zsetConvert(zobj,OBJ_ENCODING_SKIPLIST);
```