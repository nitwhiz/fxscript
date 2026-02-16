var A

macro m1
  set A, 10
endmacro

macro m2
  m1
  set A, 20
endmacro

macro m3
  m2
  m1
endmacro

_main:

  call test
  exit

test:
  set A, 30
  m1
  m2
  m3
  ret
