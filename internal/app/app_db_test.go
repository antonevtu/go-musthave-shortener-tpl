package app

//func TestDBPing(t *testing.T) {
//	var cfgApp = config(t)
//	err := db.Pool.New(context.Background(), cfgApp.DatabaseDSN)
//	assert.Equal(t, err, nil)
//	r := handlers.NewRouter(&db.Pool, cfgApp)
//	ts := httptest.NewServer(r)
//	defer ts.Close()
//
//	client := &http.Client{}
//	req, err := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
//	require.NoError(t, err)
//	resp, err := client.Do(req)
//	require.NoError(t, err)
//	require.Equal(t, resp.StatusCode, http.StatusOK)
//	defer resp.Body.Close()
//
//}

//func TestDBJSONAPI(t *testing.T) {
//	cfgApp := config(t)
//	//_ = os.Remove(cfgApp.FileStoragePath)
//
//	err := db.Pool.New(context.Background(), cfgApp.DatabaseDSN)
//	assert.Equal(t, err, nil)
//
//	r := handlers.NewRouter(&db.Pool, cfgApp)
//	ts := httptest.NewServer(r)
//	defer ts.Close()
//
//	// Create ID1
//	longURL := "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98"
//	buf := testEncodeJSONLongURL(longURL)
//	resp, shortURLInJSON := testRequest(t, ts.URL+"/api/shorten", "POST", buf)
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//	// Parse shortURL
//	shortURL := testDecodeJSONShortURL(t, shortURLInJSON)
//	_, err = url.Parse(shortURL)
//	require.NoError(t, err)
//
//	// Create ID2
//	longURL = "https://habr.com/ru/all/"
//	buf = testEncodeJSONLongURL(longURL)
//	resp, shortURLInJSON = testRequest(t, ts.URL+"/api/shorten", "POST", buf)
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//	// Parse shortURL
//	shortURL = testDecodeJSONShortURL(t, shortURLInJSON)
//	u, err := url.Parse(shortURL)
//	require.NoError(t, err)
//
//	// Check redirection by existing ID
//	resp, _ = testRequest(t, ts.URL+u.Path, "GET", nil)
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
//	longURLRecovered := resp.Header.Get("Location")
//	assert.Equal(t, longURL, longURLRecovered)
//
//	// Check StatusBadRequest for incorrect JSON key in request
//	badJSON := `{"urlBad":"abc"}`
//	resp, _ = testRequest(t, ts.URL+"/api/shorten", "POST", bytes.NewBufferString(badJSON))
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
//
//	// Check server returns correct baseURL
//	baseURL := u.Scheme + "://" + u.Host
//	assert.Equal(t, baseURL, cfgApp.BaseURL)
//
//	// Check storage file created
//	file, err := os.Open(cfgApp.FileStoragePath)
//	assert.Equal(t, err, nil)
//	defer file.Close()
//}

//func TestDBCookie(t *testing.T) {
//	cfgApp := config(t)
//	//_ = os.Remove(cfgApp.FileStoragePath)
//	err := db.Pool.New(context.Background(), cfgApp.DatabaseDSN)
//	assert.Equal(t, err, nil)
//
//	r := handlers.NewRouter(&db.Pool, cfgApp)
//	ts := httptest.NewServer(r)
//	defer ts.Close()
//
//	// Create ID1
//	longURL := "https://yandex.ru/maps/geo/sochi/53166566/?ll=39.580041%2C43.713351&z=9.98"
//	buf := testEncodeJSONLongURL(longURL)
//	resp, shortURLInJSON := testGZipRequest(t, ts.URL+"/api/shorten", "POST", buf)
//	cookies := resp.Cookies()
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//	// Check response header
//	val := resp.Header.Get("Content-Encoding")
//	assert.Equal(t, val, "gzip")
//
//	// Parse shortURL
//	shortURL := testDecodeJSONShortURL(t, shortURLInJSON)
//	_, err = url.Parse(shortURL)
//	require.NoError(t, err)
//
//	// Create ID2
//	longURL = "https://habr.com/ru/all/"
//	buf = testEncodeJSONLongURL(longURL)
//	resp, shortURLInJSON = testGZipRequestCookie(t, ts.URL+"/api/shorten", "POST", buf, cookies)
//	//cookies = resp.Cookies()
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	assert.Equal(t, http.StatusCreated, resp.StatusCode)
//
//	// Check response header
//	val = resp.Header.Get("Content-Encoding")
//	assert.Equal(t, val, "gzip")
//
//	// Parse shortURL
//	shortURL = testDecodeJSONShortURL(t, shortURLInJSON)
//	_, err = url.Parse(shortURL)
//	require.NoError(t, err)
//
//	// Test User history with true cookie
//	resp, urls := testGZipRequestCookie(t, ts.URL+"/user/urls", "GET", buf, cookies)
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	require.Equal(t, http.StatusOK, resp.StatusCode)
//	require.Equal(t, len(urls) > 2, true)
//
//	// Test User history with false cookie
//	cookies[0].Value = "4a03f9cb72d311ec9a6930c9abdb0255e0706164f46eddd837acfabfdc37ffb65275b7ca3f1e7cb9e113f4c79c33e910"
//	resp = testGZipRequestCookie204(t, ts.URL+"/user/urls", "GET", buf, cookies)
//	err = resp.Body.Close()
//	require.NoError(t, err)
//	require.Equal(t, http.StatusNoContent, resp.StatusCode)
//}