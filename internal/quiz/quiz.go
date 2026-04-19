package quiz

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/JesterForAll/gonote/internal/database"
	"github.com/JesterForAll/gonote/internal/middleware"
)

type Quiz struct {
	Logger   *slog.Logger
	DB       *database.Database
	fileList []os.DirEntry
}

type note struct {
	Note     string
	Octave   string
	AudioUrl string
}

type confirm struct {
	Correct     bool
	CorrectNote string
	Accuracy    float32
}

func newQuiz(logger *slog.Logger) (*Quiz, error) {
	fileList, err := os.ReadDir("../../assets")
	if err != nil {
		logger.Error("error getting data from disk", slog.Any("err", err))

		return nil, err
	}

	db, err := database.New("../../static/accuracy.db", logger, database.AccuracyDbStruct{})
	if err != nil {
		logger.Error("failed to connect database", slog.Any("err", err))

		return nil, err
	}

	return &Quiz{fileList: fileList, DB: db, Logger: logger}, nil
}

func (quiz *Quiz) getRandomNote() *note {

	newR := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNote := quiz.fileList[newR.Intn(len(quiz.fileList))]

	// note name is in format "1-я октава (1), ре, D.wav"
	var note note

	indxLead := strings.Index(randNote.Name(), "(")
	indxEnd := strings.Index(randNote.Name(), ")")

	if indxLead != -1 && indxEnd != -1 {
		note.Octave = randNote.Name()[indxLead+1 : indxEnd]
	}

	indxDot := strings.Index(randNote.Name(), ".")
	if indxDot != -1 {
		note.Note = strings.TrimSpace(randNote.Name()[indxEnd+2 : indxDot])
	}

	note.AudioUrl = randNote.Name()

	return &note
}

func (quiz *Quiz) processConfirmation(confirmRequest *confirmRequest, ctx context.Context) (*confirm, error) {
	var confirm confirm
	var noteData database.AccuracyDbStruct

	userID := ctx.Value(middleware.UserIDKey).(int)

	exist := quiz.DB.CheckIfExistAndGetFirst(map[string]interface{}{
		"note":    confirmRequest.CurrentNote,
		"octave":  confirmRequest.CurrentOctave,
		"user_id": userID,
	}, &noteData)

	if !exist {
		noteData.Note = confirmRequest.CurrentNote
		noteData.Octave = confirmRequest.CurrentOctave
		noteData.UserID = userID
	}

	if confirmRequest.CurrentNote == confirmRequest.Note && confirmRequest.CurrentOctave == confirmRequest.Octave {
		confirm.Correct = true
		noteData.CorrectCount++
	}

	if !confirm.Correct {
		confirm.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote
	}

	noteData.NumTries++
	noteData.Accuracy = float32(noteData.CorrectCount) / float32(noteData.NumTries) * 100.00

	err := quiz.DB.Upsert(&noteData)
	if err != nil {
		quiz.Logger.Error("error getting data from disk", slog.Any("err", err))

		return nil, err
	}

	confirm.Accuracy = noteData.Accuracy

	return &confirm, nil
}
