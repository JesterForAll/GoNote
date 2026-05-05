package quiz

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/JesterForAll/gonote/internal/balance"
	"github.com/JesterForAll/gonote/internal/database"
	"github.com/JesterForAll/gonote/internal/inventory"
	"github.com/JesterForAll/gonote/internal/transaction"
	"github.com/JesterForAll/gonote/internal/utils"
)

type Quiz struct {
	Logger    *slog.Logger
	DB        *database.Database
	fileList  []os.DirEntry
	balance   *balance.Balance
	inventory *inventory.Inventory
}

type note struct {
	Note     string
	Octave   string
	AudioURL string
}

type confirm struct {
	Correct     bool
	CorrectNote string
	Accuracy    float32
}

func newQuiz(logger *slog.Logger, balance *balance.Balance, inv *inventory.Inventory) (*Quiz, error) {
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

	return &Quiz{fileList: fileList, DB: db, Logger: logger, balance: balance, inventory: inv}, nil
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

	note.AudioURL = randNote.Name()

	return &note
}

func (quiz *Quiz) processConfirmation(ctx context.Context, confirmRequest *confirmRequest) (*confirm, error) {
	var (
		confirm  confirm
		noteData database.AccuracyDbStruct
	)

	userID, err := utils.GetUserIDFromContext(ctx, quiz.Logger)
	if err != nil {
		quiz.Logger.Error("error getting user ID from context", slog.Any("err", err))
		return nil, err
	}

	exist := quiz.DB.CheckIfExistAndGetFirst(map[string]interface{}{
		"note":    confirmRequest.CurrentNote,
		"octave":  confirmRequest.CurrentOctave,
		"user_id": userID,
	}, &noteData)

	if !exist {
		noteData.Note = confirmRequest.CurrentNote
		noteData.Octave = confirmRequest.CurrentOctave
		noteData.UserID = userID
		noteData.Accuracy = 0
	}

	if confirmRequest.CurrentNote == confirmRequest.Note && confirmRequest.CurrentOctave == confirmRequest.Octave {
		confirm.Correct = true
		noteData.CorrectCount++
	}

	valUpdateBalance := balance.BalancePlusForWin

	if !confirm.Correct {
		confirm.CorrectNote = confirmRequest.CurrentOctave + ", " + confirmRequest.CurrentNote

		valUpdateBalance *= -1
	}

	numOfSafeFails := quiz.inventory.GetCurrentNumOfSafeFails(userID)

	// we have safe fail, we need to change number of safe fails and return data, but dont change accuracy and num tries
	if numOfSafeFails > 0 && !confirm.Correct {
		_, err := quiz.inventory.UpdateCurrentNumOfSafeFails(userID, false)
		if err != nil {
			quiz.Logger.Error("error updating number of safe fails", slog.Any("err", err))

			return nil, err
		}

		confirm.Accuracy = noteData.Accuracy

		return &confirm, nil
	}

	noteData.NumTries++
	noteData.Accuracy = float32(noteData.CorrectCount) / float32(noteData.NumTries) * 100.00

	confirm.Accuracy = noteData.Accuracy

	// running a transaction
	err = transaction.RunMulti(ctx, transaction.MultiConfig{
		Name:   "process confirmation transaction",
		Logger: quiz.Logger,
		DBs:    []*database.Database{quiz.DB, quiz.balance.Db}},
		func(ctx context.Context, txs ...*gorm.DB) error {
			if err := quiz.DB.UpsertWithTx(txs[0], &noteData); err != nil {
				return fmt.Errorf("error upserting note data: %w", err)
			}

			if err := quiz.balance.UpdateCurrentBalanceWithTx(txs[1], userID, valUpdateBalance); err != nil {
				return fmt.Errorf("error updating balance: %w", err)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	return &confirm, nil
}
