package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetListOfParamsNames(t *testing.T) {
	script := `#!/busybox/sh

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

	paramsList := GetListOfParamsNames(script)

	assert.Equal(t, 7, len(paramsList))
	assert.Equal(t, "${git_repo_src}", paramsList["git_repo_src"])
	assert.Equal(t, "${DockerFileAppendCert}", paramsList["DockerFileAppendCert"])
	assert.Equal(t, "${DockerFileContext}", paramsList["DockerFileContext"])
	assert.Equal(t, "${RootCert}", paramsList["RootCert"])
	assert.Equal(t, "${DockerFilePath}", paramsList["DockerFilePath"])
	assert.Equal(t, "${str_dest}", paramsList["str_dest"])
	assert.Equal(t, "${str_kaniko_args}", paramsList["str_kaniko_args"])
}
