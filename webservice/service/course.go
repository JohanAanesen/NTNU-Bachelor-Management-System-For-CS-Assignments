package service

import (
	"database/sql"
	"errors"
	"github.com/JohanAanesen/CSAMS/webservice/model"
	"github.com/JohanAanesen/CSAMS/webservice/repository"
)

var (
	// ErrUserAlreadyInCourse error
	ErrUserAlreadyInCourse = errors.New("user already in course")
)

// CourseService struct
type CourseService struct {
	courseRepo *repository.CourseRepository
}

// NewCourseService returns a pointer to a new CourseService
func NewCourseService(db *sql.DB) *CourseService {
	return &CourseService{
		courseRepo: repository.NewCourseRepository(db),
	}
}

// Fetch a single course
func (s *CourseService) Fetch(id int) (*model.Course, error) {
	return s.courseRepo.Fetch(id)
}

// FetchAll all courses
func (s *CourseService) FetchAll() ([]*model.Course, error) {
	return s.courseRepo.FetchAll()
}

// FetchAllForUser all courses for a given user
func (s *CourseService) FetchAllForUser(userID int) ([]*model.Course, error) {
	return s.courseRepo.FetchAllForUser(userID)
}

// FetchAllForUserOrdered all courses in order for a given user
func (s *CourseService) FetchAllForUserOrdered(userID int) ([]*model.Course, error) {
	return s.courseRepo.FetchAllForUserOrdered(userID)
}

// FetchAllStudentsFromCourse all students from a course
func (s *CourseService) FetchAllStudentsFromCourse(courseID int) ([]*model.User, error) {
	return s.courseRepo.FetchAllStudentsFromCourse(courseID)
}

// Exists checks if a course exists
func (s *CourseService) Exists(hash string) *model.Course {
	result := model.Course{
		ID: -1,
	}

	courses, err := s.courseRepo.FetchAll()
	if err != nil {
		return &result
	}

	for _, course := range courses {
		if course.Hash == hash {
			return course
		}
	}

	return &result
}

// UserInCourse checks if user is in given course
func (s *CourseService) UserInCourse(userID, courseID int) (bool, error) {
	return s.courseRepo.UserInCourse(userID, courseID)
}

// AddUser to a single course
func (s *CourseService) AddUser(userID, courseID int) error {
	// Check if user already in course
	exists, err := s.UserInCourse(userID, courseID)
	if err != nil {
		return err
	}

	// Return predefined error if user in course
	if exists {
		return ErrUserAlreadyInCourse
	}

	// Add user to course
	return s.courseRepo.InsertUser(userID, courseID)
}

// RemoveUser from a course
func (s *CourseService) RemoveUser(userID, courseID int) error {
	return s.courseRepo.RemoveUser(userID, courseID)
}

// Insert course into the database
func (s *CourseService) Insert(course model.Course) (int, error) {
	return s.courseRepo.Insert(course)
}

// Update a course in the database
func (s *CourseService) Update(course model.Course) error {
	return s.courseRepo.Update(course)
}

// Delete a course in the database
func (s *CourseService) Delete(id int) error {
	return s.courseRepo.Delete(id)
}
