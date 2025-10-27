# Warning
Использование -B, -A, -C, -n для мультисерверного режима недоступно
Тк мы используем шардирование и из-за недостатка контекста
не имеем возможность определить его для совпавшей строки

# Comparing
Сравнение производилось на файле размером 205MB(1.891.715 строк)
patterns:
- ppp224.st.rim.or.jp (23 in result)
- DELETE (0 in result)
- '^.*$' (1.891.715 in result)
При помощи команды /usr/bin/time

## Default GREP
```
command: /usr/bin/time -al grep ppp224.st.rim.or.jp access.log
time: 1,02 real         0,91 user         0,03 sys
memory: ≈1.4MB

command: /usr/bin/time -al grep DELETE access.log
time: 0,30 real         0,27 user         0,02 sys
memory: ≈1.3MB

command: /usr/bin/time -al grep '^.*$' access.log
time: 9,09 real         4,60 user         1,46 sys
memory: ≈15MB
```

## MyGrep MutiServer mode 
```
command: /usr/bin/time -al ./mygrep DELETE -f access.log -m
time: 0,57 real         1,51 user         1,03 sys
memory: ≈15MB

command: /usr/bin/time -al ./mygrep ppp224.st.rim.or.jp -f access.log -m
time: 0,66 real         1,75 user         1,08 sys
memory: ≈17.58MB


command: /usr/bin/time -al ./mygrep '^.*$' -f access.log -m
time: 8,39 real        27,58 user         3,19 sys
memory: ≈390MB
```

## MyGrep Default mode
```
command: /usr/bin/time -al ./mygrep DELETE -f access.log
time: 0,40 real         0,23 user         0,06 sys
memory: ≈10.65MB

command: /usr/bin/time -al ./mygrep ppp224.st.rim.or.jp -f access.log 
time: 0,39 real         0,30 user         0,05 sys
memory: ≈10.65MB

command: /usr/bin/time -al ./mygrep '^.*$' -f access.log
time: 7,19 real         3,10 user         1,72 sys
memory: ≈453MB

# without vectoring out strings, only in default mode
command: /usr/bin/time -al ./mygrep '^.*$' -f access.log -V
time: 5,94 real         2,60 user         1,34 sys
memory: ≈10MB
```

# Ограничения для МультиСерверного режима
- AfterMatch (-A) = 0
- BeforeMatch (-B) = 0
- AroundMatch (-C) = 0
- NumberForStringsFlag (-n) = false

Эти ограничения связаны с тем, что мы используем шардирование,
cл-но данные разбиваются на части и мы теряем контекст некоторых строк,
которые остаются на границах, также теряем возможность отслеживать номер 
строки в файле, тк каждая нода читает "свой вариант" файла

# Использование памяти
Такой большой объем использованной памяти связан с тем, что для каждого файла
мы формируем вектор со строками удовлятворяющими условию, есть возможность
отказаться от этого, но не в мультисерверном режиме, тк нам необходимо вначале
отсортировать наши шарды в нужном порядке по их ID и дождаться согласования кворума по каждому, что вынуждает нас хранить эти данные некоторое время