from fabric import Connection, task
from invoke import run

@task
def deploy(c):
    with Connection('hongkong','root') as c:
        c.run("sudo systemctl stop red-or-black", pty=True)

    run("export GOPROXY=https://goproxy.cn && env GOOS=linux GOARCH=amd64 go build -o red-or-black")
    run("scp red-or-black hongkong:/usr/local/red-or-black")
    run("scp red-or-black.service hongkong:/etc/systemd/system")

    with Connection('hongkong','root') as c:
        c.run("sudo systemctl start red-or-black", pty=True)