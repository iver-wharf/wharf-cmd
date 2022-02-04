package kaniko

import "strings"

const script = `#!/busybox/sh

cp -R ${git_repo_src}/* 'pwd'

if [ "${DockerFileAppendCert}" == "true" ]
then

cert_path='pwd'/${DockerFileContext}/root.crt

cat <<EOF > root.crt
${RootCert}
EOF

cp root.crt $cert_path

echo -e "\n" >> 'pwd'/${DockerFilePath}
echo COPY ./root.crt /usr/local/share/ca-certificates/root.crt >> 'pwd'/${DockerFilePath}
echo "RUN mkdir -p /etc/ssl/certs/ && touch /etc/ssl/certs/ca-certificates.crt && cat /usr/local/share/ca-certificates/root.crt >> /etc/ssl/certs/ca-certificates.crt"  >> 'pwd'/${DockerFilePath}

fi

cat 'pwd'/${DockerFilePath}

set -x
/kaniko/executor -f 'pwd'/${DockerFilePath} -c 'pwd' --context=dir://'pwd'/${DockerFileContext} --skip-tls-verify ${str_dest} --verbosity warn ${str_kaniko_args}`

func GetScript() string {
	return strings.Replace(script, "'", "`", -1)
}
