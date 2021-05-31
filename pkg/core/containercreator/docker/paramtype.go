package docker

const (
	kanikoArgs = "str_kaniko_args"
	stringDest = "str_dest"
	gitRepoSrc = "git_repo_src"
)

type ParamType int

const (
	FileArgs = ParamType(iota)
	FileAppendCert
	FileContext
	FilePath
	ImageDestination
	ImageTags
	ImagePush
	Insecure
	RegSecret
	RootCert
	KanikoArgs
	StringDest
	GitRepoSrc
)

var toPT = map[string]ParamType{
	"DockerFileArgs":    FileArgs,
	"FileAppendCert":    FileAppendCert,
	"DockerFileContext": FileContext,
	"DockerFilePath":    FilePath,
	"ImageDestination":  ImageDestination,
	"ImageTags":         ImageTags,
	"ImagePush":         ImagePush,
	"Insecure":          Insecure,
	"RegSecret":         RegSecret,
	"RootCert":          RootCert,
	kanikoArgs:          KanikoArgs,
	stringDest:          StringDest,
	gitRepoSrc:          GitRepoSrc,
}

var ptToStr = map[ParamType]string{}

func init() {
	for str, st := range toPT {
		ptToStr[st] = str
	}
}

func (t ParamType) String() string {
	return ptToStr[t]
}

func ParseParam(name string) ParamType {
	return toPT[name]
}
