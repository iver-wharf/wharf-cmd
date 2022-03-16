package wharfyml

func listOSPossibleVarsFiles() []VarFile {
	return []VarFile{
		{
			Path: "/etc/iver-wharf/wharf-cmd/" + builtInVarsFile,
			Kind: VarFileKindConfigDir,
		},
	}
}
