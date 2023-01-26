## GoShellSec

## Описание
Ansible-like инструмент для асинхронного выполнения удаленной команды через протокол SSH на нескольких серверах. 

*Отключена строгая проверка ключа*
## Использование
```bash
$ ./goshell --help
Usage of ./goshell:
  --cmd string
    	Команда которая должна выполниться (default "uname -a")
  --file string
    	Файл с нужными адресами
  --hosts string
        Адреса через запятую, чтобы не использовать файл
  --grep string
    	Поиск совпадений в выводе
  --pass string
    	Пароль для входа
  --port string
    	SSH Порт (default "22")
  --sshname string
      Имя приватного ключа (имя файла, который лежит в ~/.ssh/) (default "id_rsa")
  --timeout string
      Время выполнения (ex: 10s, 1m, 2h) (default "5s")
  --user string
    	Пользователь для входа
```

## Зависимости
* go
* docker*

## Собираем проект (любая ОС)
```bash
$ go build -o goshell cmd/main.go
$ goshell --help
```

## Для старых версий ОС (centos 7)
```bash
$ docker run --rm -it -v $(pwd):/opt centos:7
$ (centos7) $ wget https://dl.google.com/go/go1.13.linux-amd64.tar.gz
$ (centos7) $ tar -C /usr/local -xzf go1.13.linux-amd64.tar.gz
$ (centos7) $ export PATH=$PATH:/usr/local/go/bin
$ (centos7) $ cd /opt; ls; go version
$ (centos7) $ go build -o goshell cmd/main.go
```

## Примеры
```bash
goshell --cmd "uname -r" --file hosts --user test          # Выполняет команду используя приватный ключ для входа
goshell --cmd "uname -r" --port 3022 --file hosts --user test --pass "password123" # Используя пароль и порт 3022 
```