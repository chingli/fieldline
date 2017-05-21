/*
tensor2d 包实现了一个定义在二维笛卡尔坐标系内的对称 2 阶张量.
通常用此张量来表示应力或应变.

该张量的表示形式为:

	      ┌ XX  XY ┐
	Tij = │        │
	      └ YX  YY ┘

根据剪应力/剪应变互等定理, 对于此 2 阶张量, 以下关系总成立:

	XY = YX

由于张量矩阵为实对称矩阵, 因此其特征值总为实数, 且不论特征值是否相重,
总存在一组完整的标准正交特征向量.

一个 Tensor2d 可以分解为 2 个 vector.
*/
package tensor
