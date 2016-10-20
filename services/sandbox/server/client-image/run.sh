if [ -f "__init__.py" ]; then
    python3 __init__.py
elif [ -f "main.go" ]; then
    go run main.go
else
    echo "No script found to run."
fi
