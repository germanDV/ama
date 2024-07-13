package questionnaire

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepo(host string, port int, password string, ttl time.Duration) Repository {
	return &RedisRepository{
		ttl: ttl,
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       0,
		}),
	}
}

func (r *RedisRepository) SaveQuestionnaire(q Questionnaire) error {
	key := fmt.Sprintf("QA:%s", q.ID)
	val, err := json.Marshal(q)
	if err != nil {
		return err
	}
	return r.client.Set(context.TODO(), key, val, r.ttl).Err()
}

func (r *RedisRepository) GetQuestionnaire(questionnaireID string) (Questionnaire, error) {
	key := fmt.Sprintf("QA:%s", questionnaireID)
	q := Questionnaire{}

	val, err := r.client.Get(context.TODO(), key).Bytes()
	if err != nil {
		return Questionnaire{}, err
	}

	err = json.Unmarshal(val, &q)
	if err != nil {
		return Questionnaire{}, err
	}

	return q, nil
}

func (r *RedisRepository) CountQuestionnaires() (int, error) {
	keys, cursor, err := r.client.Scan(context.TODO(), 0, "QA:*", 100).Result()
	if err != nil {
		return 0, err
	}
	if cursor != 0 {
		return 0, fmt.Errorf("more than 100 questionnaires")
	}
	return len(keys), nil
}

func (r *RedisRepository) SaveQuestion(questionnaireID string, q Question) error {
	key := fmt.Sprintf("%s:%s", questionnaireID, q.ID)
	val, err := json.Marshal(q)
	if err != nil {
		return err
	}
	return r.client.Set(context.TODO(), key, val, r.ttl).Err()
}

func (r *RedisRepository) GetQuestions(questionnaireID string) ([]Question, error) {
	keyPattern := fmt.Sprintf("%s:*", questionnaireID)
	keys, cursor, err := r.client.Scan(context.TODO(), 0, keyPattern, 100).Result()
	if err != nil {
		return nil, err
	}
	if cursor != 0 {
		return nil, fmt.Errorf("more than 100 questions in questionnaire %s", questionnaireID)
	}

	qs := make([]Question, 0, len(keys))
	for _, key := range keys {
		val, err := r.client.Get(context.TODO(), key).Bytes()
		if err != nil {
			return nil, err
		}

		q := Question{}
		err = json.Unmarshal(val, &q)
		if err != nil {
			return nil, err
		}

		qs = append(qs, q)
	}

	return qs, nil
}

func (r *RedisRepository) CountQuestions(questionnaireID string) (int, error) {
	keyPattern := fmt.Sprintf("%s:*", questionnaireID)
	keys, cursor, err := r.client.Scan(context.TODO(), 0, keyPattern, 100).Result()
	if err != nil {
		return 0, err
	}
	if cursor != 0 {
		return 0, fmt.Errorf("more than 100 questions in questionnaire %s", questionnaireID)
	}
	return len(keys), nil
}

func (r *RedisRepository) updateQuestion(key string, kind string) (Question, error) {
	val, err := r.client.Get(context.TODO(), key).Bytes()
	if err != nil {
		return Question{}, err
	}

	q := Question{}

	err = json.Unmarshal(val, &q)
	if err != nil {
		return Question{}, err
	}

	if kind == "vote" {
		q.Metadata.Votes++
	} else if kind == "answer" {
		q.Metadata.Answered = true
	} else {
		return Question{}, fmt.Errorf("unknown update kind: %s", kind)
	}

	val, err = json.Marshal(q)
	if err != nil {
		return Question{}, err
	}

	exp, err := r.client.ExpireTime(context.Background(), key).Result()
	if err != nil {
		return Question{}, err
	}

	err = r.client.Set(context.TODO(), key, val, exp).Err()
	if err != nil {
		return Question{}, err
	}

	return q, nil
}

func (r *RedisRepository) Vote(questionnaireID string, questionID string) (uint16, error) {
	key := fmt.Sprintf("%s:%s", questionnaireID, questionID)
	updatedQ, err := r.updateQuestion(key, "vote")
	if err != nil {
		return 0, err
	}
	return updatedQ.Metadata.Votes, nil
}

func (r *RedisRepository) Answer(questionnaireID string, questionID string) error {
	key := fmt.Sprintf("%s:%s", questionnaireID, questionID)
	_, err := r.updateQuestion(key, "answer")
	return err
}
