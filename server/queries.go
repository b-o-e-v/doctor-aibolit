package server

// запросы в базу
const QUERY_USER_EXISTS = `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`
const QUERY_USER_CREATE = `INSERT INTO users (id, name) VALUES ($1, $2) RETURNING id`
const QUERY_MEDICATION_EXISTS = `SELECT EXISTS(SELECT 1 FROM medications WHERE id=$1)`
const QUERY_MEDICATION_CREATE = `INSERT INTO medications (id, name) VALUES ($1, $2) RETURNING id`
const QUERY_TAKING_CREATE = `INSERT INTO takings (schedule_id, taking_time) VALUES ($1, $2)`
const QUERY_SCHEDULE_CREATE = `
  INSERT INTO schedules (user_id, medication_id, frequency, duration, start_date)
  VALUES ($1, $2, $3, $4, NOW())
  RETURNING id, start_date, end_date, EXTRACT(EPOCH FROM frequency) AS frequency_seconds
`
const QUERY_USER_SCHEDULES = `SELECT id FROM schedules WHERE user_id = $1`
const QUERY_USER_SCHEDULE = `
	SELECT t.id, t.taking_time, s.medication_id
	FROM takings AS t
	JOIN schedules AS s ON t.schedule_id = s.id
	WHERE t.taking_time >= NOW() 
		AND t.taking_time::DATE = CURRENT_DATE
		AND s.user_id = $1
		AND t.schedule_id = $2
	ORDER BY t.taking_time ASC
`
const QUERY_USER_TAKINGS = `
	SELECT t.id, t.taking_time, s.medication_id, s.id
	FROM takings AS t
	JOIN schedules AS s ON t.schedule_id = s.id
	WHERE t.taking_time >= NOW()
		AND t.taking_time <= NOW() + $1::INTERVAL
		AND s.user_id = $2
	ORDER BY t.taking_time ASC
`
