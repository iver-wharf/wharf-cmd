package containercreator

type InitContainerType uint

const (
	Git = InitContainerType(1 << iota)
)

const AllAvailableInitContainers = InitContainerType(0x1)

func (c InitContainerType) IsValid() bool {
	return AllAvailableInitContainers&c == c
}

func (c InitContainerType) toPodContainersFlag() PodContainersFlags {
	return PodContainersFlags(c)
}

type ContainerType uint

const (
	Container = ContainerType(1 << iota)
	Kaniko
	Docker
	Helm
	KubeApply
)

const AllAvailableContainers = ContainerType(0x1F)

func (c ContainerType) IsValid() bool {
	return AllAvailableContainers&c == c
}

func (c ContainerType) toPodContainersFlag() PodContainersFlags {
	return PodContainersFlags(c << 32)
}

type PodContainersFlags uint64

const (
	initContainersMask = PodContainersFlags(0x00000000FFFFFFFF)
	containersMask     = PodContainersFlags(0xFFFFFFFF00000000)
)

func (p PodContainersFlags) AddContainer(c ContainerType) PodContainersFlags {
	return p | (c.toPodContainersFlag() & containersMask)
}

func (p PodContainersFlags) AddInitContainer(c InitContainerType) PodContainersFlags {
	return p | (c.toPodContainersFlag() & initContainersMask)
}

func (p PodContainersFlags) GetInitContainersCount() uint {
	return bitsCounter(uint(p.GetInitContainers()), uint(AllAvailableInitContainers))
}

func (p PodContainersFlags) GetInitContainers() InitContainerType {
	return InitContainerType(p & initContainersMask)
}

func (p PodContainersFlags) GetContainersCount() uint {
	return bitsCounter(uint(p.GetContainers()), uint(AllAvailableContainers))
}

func (p PodContainersFlags) GetContainers() ContainerType {
	return ContainerType((p & containersMask) >> 32)
}

func (p PodContainersFlags) HasContainer(c ContainerType) bool {
	return p.GetContainers()&c == c
}

func (p PodContainersFlags) HasInitContainer(c InitContainerType) bool {
	return p.GetInitContainers()&c == c
}

func bitsCounter(flags uint, maxValue uint) uint {
	validContainers := flags & maxValue

	testMask := uint(0x1)
	count := uint(0)
	for testMask <= maxValue {
		if validContainers&testMask == testMask {
			count++
		}
		testMask = testMask << 1
	}
	return count
}
