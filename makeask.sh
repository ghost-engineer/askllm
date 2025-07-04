#!/bin/bash

# 👉 add to ~/.bashrc
echo '
# 🧠 ask for easy curl-query
ask() {
  query="$*"
  curl -G --data-urlencode "q=${query}" http://localhost:8080/
}

# 🚀 token for API
export CHUTES_API_TOKEN="your_token_here"
' >> ~/.bashrc

# ✅ apply changes
source ~/.bashrc

# 🎉 greetings 
echo "ask & CHUTES_API_TOKEN ready."