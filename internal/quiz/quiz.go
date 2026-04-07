package quiz

import (
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/JesterForAll/gonote/internal/database"
)

type Quiz struct {
	Logger   *slog.Logger
	DB       *database.Database
	fileList []os.DirEntry
}

type note struct {
	Note     string `json:"note"`
	Octave   string `json:"octave"`
	Index    int    `json:"index"`
	AudioUrl string `json:"audioUrl"`
}

func newQuiz(logger *slog.Logger) (*Quiz, error) {
	fileList, err := os.ReadDir("../../assets")
	if err != nil {
		logger.Error("error getting data from disk", slog.Any("err", err))

		return nil, err
	}

	db, err := database.New("../../static/accuracy.db")
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
	var NoteForSrv note

	indxLead := strings.Index(randNote.Name(), "(")
	indxEnd := strings.Index(randNote.Name(), ")")

	if indxLead != -1 && indxEnd != -1 {
		NoteForSrv.Octave = randNote.Name()[indxLead+1 : indxEnd]
	}

	indxDot := strings.Index(randNote.Name(), ".")
	if indxDot != -1 {
		NoteForSrv.Note = strings.TrimSpace(randNote.Name()[indxEnd+2 : indxDot])
	}

	NoteForSrv.AudioUrl = randNote.Name()

	return &NoteForSrv
}

func (quiz *Quiz) processConfirmation(confirmRequest *confirmRequest) *confirmResponse {
	var confRes confirmResponse
	var noteData database.DbStruct

	res := quiz.DB.Get(map[string]interface{}{"Note": confirmRequest.CurrentNote, "Octave": confirmRequest.CurrentOctave}, &noteData)

	if !res {
		noteData.Note = confirmRequest.CurrentNote
		noteData.Octave = confirmRequest.CurrentOctave
	}

	if confirmRequest.CurrentNote == confirmRequest.Note && confirmRequest.CurrentOctave == confirmRequest.Octave {
		confRes.Correct = true
		noteData.CorrectCount++
	}

	if !confRes.Correct {
		confRes.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote
	}

	noteData.NumTries++
	noteData.Accuracy = float32(noteData.CorrectCount) / float32(noteData.NumTries) * 100.00

	//update data in db
	quiz.DB.Save(&noteData)

	confRes.Accuracy = noteData.Accuracy

	return &confRes
}
