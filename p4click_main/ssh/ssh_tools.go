package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	. "p4click/types"
	"time"
)

// link: http://blog.ralch.com/tutorial/golang-ssh-connection/


type SSH struct{
	Ip string
	User string
	Cert string //password or key file path
	Port int
	session *ssh.Session
	client *ssh.Client
}

func (ssh_client *SSH) readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (ssh_client *SSH) Connect(mode int) {

	var ssh_config *ssh.ClientConfig
	var auth  []ssh.AuthMethod
	if mode == CERT_PASSWORD {
		auth = []ssh.AuthMethod{ssh.Password(ssh_client.Cert)}
	} else if mode == CERT_PUBLIC_KEY_FILE {
		auth = []ssh.AuthMethod{ssh_client.readPublicKeyFile(ssh_client.Cert)}
	} else {
		log.Println("does not support mode: ", mode)
		return
	}

	ssh_config = &ssh.ClientConfig{
		User: ssh_client.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout:time.Second * DEFAULT_TIMEOUT,
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ssh_client.Ip, ssh_client.Port), ssh_config)
	if err != nil {
		fmt.Println(err)
		return
	}

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		client.Close()
		return
	}

	ssh_client.session = session
	ssh_client.client  = client
}

func (ssh_client *SSH) GetCmdOutput(cmd string) (string, error) {
	out, err := ssh_client.session.Output(cmd)
	return string(out), err
}

func (ssh_client *SSH) Close() {
	ssh_client.session.Close()
	ssh_client.client.Close()
}