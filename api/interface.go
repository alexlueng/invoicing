package api

// 学习面向接口编程
// 调用未来可能就是接口的最大意义所在吧，这也是为什么架构师那么值钱，
// 因为良好的架构师是可以针对interface设计一套框架，在未来许多年却依然适用。

type Person interface {
	List()
	Add()

}
