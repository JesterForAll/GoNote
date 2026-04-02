package quiz

import (
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/JesterForAll/gonote/internal/database"
)

type availibleNotes struct {
	Octaves []string `json:"octaves"`
	Notes   []string `json:"notes"`
}

type note struct {
	Note     string `json:"note"`
	Octave   string `json:"octave"`
	Index    int    `json:"index"`
	AudioUrl string `json:"audioUrl"`
}

type confirmResponse struct {
	Correct     bool    `json:"correct"`
	CorrectNote string  `json:"correctNote"`
	Accuracy    float32 `json:"accuracy"`
}

type confirmRequest struct {
	Note          string `json:"note"`
	Octave        string `json:"octave"`
	CurrentNote   string `json:"currentNote"`
	CurrentOctave string `json:"currentOctave"`
}

var ListNotes = availibleNotes{
	Octaves: []string{"-4", "-3", "-2", "-1", "1", "2", "3", "4", "5"},
	Notes:   []string{"до, C", "до#, C#", "ре, D", "ре#, D#", "ми, E", "фа, F", "фа#, F#", "соль, G", "соль#, G#", "ля, A", "ля#, A#", "си, B"},
}

type Quiz struct {
	Logger *slog.Logger
	DB     *database.Database
}

func New(logger *slog.Logger) (*Quiz, error) {

	db, err := database.New("../../static/accuracy.db")
	if err != nil {
		logger.Error("failed to connect database", slog.Any("err", err))

		return nil, err
	}

	return &Quiz{Logger: logger, DB: db}, nil
}

func (quiz *Quiz) HandleGetAvailibleNotes(w http.ResponseWriter, _ *http.Request) {

	data, err := json.Marshal(ListNotes)
	if err != nil {
		quiz.Logger.Error("error encoding data", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding data", http.StatusInternalServerError)

		return
	}

	w.Write(data)

}

func (quiz *Quiz) HandleGetNextNote(w http.ResponseWriter, _ *http.Request) {

	fileList, err := os.ReadDir("../../assets")
	if err != nil {
		quiz.Logger.Error("error getting data from disk", slog.Any("err", err))
		http.Error(w, "Internal server error while getting data from disk", http.StatusInternalServerError)

		return
	}

	newR := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNote := fileList[newR.Intn(len(fileList))]

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

	data, err := json.Marshal(NoteForSrv)
	if err != nil {
		quiz.Logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	quiz.Logger.Info("Отправлен ответ: \n", "data", data)

	w.Write(data)

}

func (quiz *Quiz) HandlePostConfirm(w http.ResponseWriter, r *http.Request) {

	var confirmRequest confirmRequest

	err := json.NewDecoder(r.Body).Decode(&confirmRequest)
	if err != nil {
		quiz.Logger.Error("error decoding request", slog.Any("err", err))
		http.Error(w, "Internal server error while decoding body", http.StatusInternalServerError)

		return
	}

	quiz.Logger.Info("got input\n", "confirmRequest", confirmRequest)

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
	} else {
		confRes.Correct = false
		confRes.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote
	}

	noteData.NumTries++
	noteData.Accuracy = float32(noteData.CorrectCount) / float32(noteData.NumTries) * 100.00

	//update data in db
	quiz.DB.Save(&noteData)

	confRes.Accuracy = noteData.Accuracy

	data, err := json.Marshal(confRes)
	if err != nil {
		quiz.Logger.Error("error encoding response", slog.Any("err", err))
		http.Error(w, "Internal server error while encoding response", http.StatusInternalServerError)

		return
	}

	quiz.Logger.Info("response\n", "data", data)

	w.Write(data)

}
