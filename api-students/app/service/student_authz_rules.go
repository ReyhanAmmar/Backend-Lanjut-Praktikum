package service

import (
	"api-students/app/model"
	"api-students/helper"
)

func CanAccessStudent(
	current model.AuthStudent,
	ownerID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.StudentID == ownerID {
		return true
	}

	return perms != nil && perms.Can(current.Role, anyPermission)
}