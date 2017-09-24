format long

ten=[20, 10; 10, 40]
[ecos,eval]=eig(ten)
ed1=acos(ecos(1,1))
ed2=asin(ecos(1,1))

ten=[20, -10; -10, 40]
[ecos,eval]=eig(ten)
ed1=acos(ecos(1,1))
ed2=asin(ecos(1,1))

ten=[40, 10; 10, 40]
[ecos,eval]=eig(ten)
ed1=acos(ecos(1,1))
ed2=asin(ecos(1,1))

ten=[40, -10; -10, 40]
[ecos,eval]=eig(ten)
ed1=acos(ecos(1,1))
ed2=asin(ecos(1,1))

ten=[40, 0; 0, -10]
[ecos,eval]=eig(ten)
ed1=acos(ecos(1,1))
ed2=asin(ecos(1,1))

ten=[40, 0; 0, 40]
[ecos,eval]=eig(ten)
ed1=acos(ecos(1,1))
ed2=asin(ecos(1,1))