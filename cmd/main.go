package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	cmd := flag.String("cmd", "uname -a", "Команда которая должна выполниться")
	file := flag.String("file", "", "Файл с нужными адресами")
	user := flag.String("user", "", "Пользователь для входа")
	pass := flag.String("pass", "", "Пароль для входа")
	port := flag.String("port", "22", "SSH Порт")
	grep := flag.String("grep", "", "Поиск совпадений в выводе")
	sshName := flag.String("sshname", "id_rsa", "Имя приватного ключа (имя файла, который лежит в ~/.ssh/)")
	someHosts := flag.String("hosts", "", "Адреса через запятую, чтобы не использовать файл")
	timeoutSec := flag.String("timeout", "10s", "Время выполнения (ex: 10s, 1m, 2h)")

	flag.Parse()

	var config *ssh.ClientConfig
	timeDuration, err := time.ParseDuration(*timeoutSec)
	if err != nil {
		log.Fatal(err)
	}
	if *user == "" {
		log.Println("--user is NULL")
		os.Exit(0)
	}
	// if *file == "" && *someHosts == "" {
	// 	log.Println("--file or --hosts are NULL")
	// 	os.Exit(0)
	// }

	if *pass == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("error get homedir: %v", err)
		}
		key, err := ioutil.ReadFile(path.Join(homeDir, ".ssh", *sshName))
		if err != nil {
			log.Fatalf("unable to read private key: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			log.Fatalf("unable to parse private key: %v", err)
		}
		config = &ssh.ClientConfig{
			User: *user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         timeDuration - (3 * time.Second),
		}
	} else {
		config = &ssh.ClientConfig{
			User: *user,
			Auth: []ssh.AuthMethod{
				ssh.Password(*pass),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         timeDuration - (3 * time.Second),
		}
	}
	hosts := []string{}
	if *file != "" {
		hosts, err = readHosts(*file)
		if err != nil {
			log.Fatal(err)
		}
	} else if *someHosts != "" {
		hosts = strings.Split(*someHosts, ",")
	} else {
		log.Fatal("--file or --hosts are NULL")
	}

	if len(hosts) < 1 {
		log.Fatal("size hosts is 0")
	}

	results := make(chan string, 10)
	timeoutGlobal := time.After(timeDuration)
	fmt.Println("Timeout Global: ", *timeoutSec)

	for _, hostname := range hosts {
		if hostname == "" {
			continue
		}
		go func(hostname string) {
			results <- executeCmd(*cmd, *grep, hostname, *port, config)
		}(hostname)
	}

	// соберем результаты со всех серверов, или напишем "Timed out", если общее время исполнения истекло
	for i := 0; i < len(hosts); i++ {
		select {
		case res := <-results:
			fmt.Println(res)
		case <-timeoutGlobal:
			fmt.Println("Done")
			return
		}
	}
}

func readHosts(pathToFile string) ([]string, error) {
	file, err := os.Open(pathToFile)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	hosts := []string{}

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		hosts = append(hosts, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}
	return hosts, nil
}

func executeCmd(cmd, grep, hostname, port string, config *ssh.ClientConfig) string {
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", hostname, port), config)
	if err != nil {
		return fmt.Sprintf("%s: %s", hostname, err.Error())
	}
	session, err := conn.NewSession()
	if err != nil {
		return fmt.Sprintf("%s: %s", hostname, err.Error())
	}
	defer session.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	if err := session.Run(cmd); err != nil {
		return fmt.Sprintf("%s: %s", hostname, err.Error())
	}
	payload := []string{}
	scanner := bufio.NewScanner(&stdoutBuf)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), grep) {
			payload = append(payload, scanner.Text())
		}
	}

	return fmt.Sprintf("%s: %s", hostname, strings.Join(payload, ","))
}
