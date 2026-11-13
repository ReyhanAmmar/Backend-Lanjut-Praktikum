package service

import (
	"errors"
	"fmt"

	"api-students/app/model"
)

var errUnknownEvent = errors.New("jenis event tidak dikenal")

func BuildMessage(eventType string, e model.EnrollmentEvent, student model.Student) (model.Message, error) {
	course := e.CourseCode + " " + e.CourseName
	var subject, body string
	switch eventType {
	case model.EventEnrollmentCreated:
		subject = "KRS: berhasil terdaftar di " + course
		body = fmt.Sprintf("Halo %s, Anda tercatat sebagai peserta kelas %s.", student.Name, course)
	case model.EventEnrollmentDropped:
		subject = "KRS: keluar dari " + course
		body = fmt.Sprintf("Halo %s, Anda sudah tidak terdaftar di kelas %s.", student.Name, course)
	case model.EventEnrollmentSwapped:
		from := e.FromCourseCode + " " + e.FromCourseName
		subject = "KRS: pindah dari " + from + " ke " + course
		body = fmt.Sprintf("Halo %s, Anda telah pindah dari kelas %s ke kelas %s.", student.Name, from, course)
	default:
		return model.Message{}, fmt.Errorf("%w: %q", errUnknownEvent, eventType)
	}
	return model.Message{To: student.Email, Subject: subject, Body: body}, nil
}

func BuildQuotaChangedMessage(
    event model.CourseQuotaChangedEvent,
    student model.Student,
) model.Message {

    course := event.CourseCode + " " + event.CourseName

    return model.Message{
        To: student.Email,
        Subject: fmt.Sprintf(
            "KRS: kuota %s berubah menjadi %d",
            course,
            event.Quota,
        ),
        Body: fmt.Sprintf(
            "Halo %s, kuota kelas %s berubah menjadi %d.",
            student.Name,
            course,
            event.Quota,
        ),
    }
}