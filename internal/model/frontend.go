package model

type Frontend struct {
	ID       FrontendID
	Name     string
	Package  PackageRef
	Launcher LauncherInfo
}
