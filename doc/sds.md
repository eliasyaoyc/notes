# SDS (Static Dynamic Strings)
简单动态字符串（Simple Dynamic Strings,SDS） 是Redis 的基本数据结构之一，用于存储字符串和整型数据。SDS兼容C语言标准字符串处理函数，且在此基础上保证了二进制安全。
> 二进制安全：C 语言中，用 " \0 " 表示字符串的结束，如果字符串中本身就有 " \0 " 字符，字符串就会被截断，即非二进制安全；
> 若通过某种机制，保证了读写字符串时不损害其内容，则是二进制安全

SDS 结构：
```c
struct sds {
  int len; 
  int free;
  char buf[];
}
```
* len 和 free 称为头部，可以方便的得到字符串长度
* 内容存放在柔性数组 buf 中
* 由于长度统计变量 len 的存在，读写字符串时不依赖 " \0 " 终止符，保证了二进制安全

> 注意：柔性数据成员（flexible array member），也叫伸缩性数组成员，只能被放在结构的末尾。
> 包含柔性数组成员的结构体，通过 malloc 函数为柔性数组动态分配内存

使用柔性数组存放字符串，是因为柔性数组的地址和结构体是连续的，这样查找内存更快（因为不需要额外通过指针字符串的位置）；
可以很方便地通过柔性数组的首地址偏移得到结构体首地址，进而能很方便地获取其余变量。

到这里已经实现了最基本的动态字符串，但是存在以下问题：
* 问题1：不同长度的字符串是否有必要占用相同的大小的头部？
* 问题2：一个int占4字节，在实际应用中，存放于Redis 中的字符串往往没有这么长，每个字符串都用4个字节存储太浪费空间
* 问题3：短字符串，len 和 free的长度为1个字节就够了；长字符串，用2字节或4字节；更长的字符串，用8字节。
> 对于以上问题，考虑增加一个字段flags 来标识类型，用最小的1字节来存储，且把flags加在柔性数组buf之前，这样虽然多了一个字节，但通过偏移柔性数组的指针即能
> 快速定位flags，区分类型，也可以接受。

结合以上问题，五种类型（长度1字节、2字节、4字节、8字节、小于1字节）的 SDS 至少要用3位来存储类型，1个字节8位，剩余5位存储长度，可以满足长度小于32的短字符串。
```c
struct __attribute__ ((__packed__)) sdshdr5 {
    unsigned char flags; // 低3位存储类型，高5位存储长度
    char buf[]; // 柔性数组，存放实际内容
};
```
!()[/sdshdr5.png]

而长度大于31的字符串，1个字节依然存不下。将 len 和 free 单独存放
```c
struct __attribute__ ((__packed__)) sdshdr8 {
    uint8_t len; // 表示 buf 中已占用字节数
    uint8_t alloc; // 表示 buf 中已分配字节数，不同于free，记录的是为buf分配的总长度
    unsigned char flags; // 标识当前结构体的类型，低3位用作标识位，高5位预留
    char buf[]; // 柔性数组，真正存储字符串的数据空间
};
struct __attribute__ ((__packed__)) sdshdr16 {
    uint16_t len; /* used */
    uint16_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr32 {
    uint32_t len; /* used */
    uint32_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr64 {
    uint64_t len; /* used */
    uint64_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
```

到此整个 sds 介绍完毕了，总结一下就是根据字符串的大小，来动态的创建对象避免小字符串和大字符串使用相同的对象导致空间的浪费，并且还使用了 `packed` 修饰词来改变字节对齐。
尽可能的不浪费一个字节。Go版本的具体实现请看 `sds.go`