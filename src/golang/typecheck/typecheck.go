package typecheck

import "golang/ast"

func CheckPackage(pkg *ast.Package) error {
	return checkPkgLevelTypes(pkg)
}
