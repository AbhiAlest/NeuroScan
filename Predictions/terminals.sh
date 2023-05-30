# Terminal 1, deployment
cd brain-cancer-detector
npm start

# Terminal 2
export FLASK_APP=app.py
export FLASK_ENV=development
flask run
