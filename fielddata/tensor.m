format long

ten=[20, 10; 10, 40]
[ecos,eval]=eig(ten)
ed1=atan(ecos(2,1)/ecos(1,1))
ed2=atan(ecos(2,2)/ecos(1,2))

ten=[20, -10; -10, 40]
[ecos,eval]=eig(ten)
ed1=atan(ecos(2,1)/ecos(1,1))
ed2=atan(ecos(2,2)/ecos(1,2))

ten=[40, 10; 10, 40]
[ecos,eval]=eig(ten)
ed1=atan(ecos(2,1)/ecos(1,1))
ed2=atan(ecos(2,2)/ecos(1,2))

ten=[40, -10; -10, 40]
[ecos,eval]=eig(ten)
ed1=atan(ecos(2,1)/ecos(1,1))
ed2=atan(ecos(2,2)/ecos(1,2))

ten=[40, 0; 0, 40]
[ecos,eval]=eig(ten)
ed1=atan(ecos(2,1)/ecos(1,1))
ed2=atan(ecos(2,2)/ecos(1,2))