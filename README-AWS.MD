
### README file focused on being easy to review 
# Logical Test


---
### Approach

The solution uses a **dynamic programming** approach optimized for **space efficiency**. Instead of maintaining a full DP array, i use two variables (`waysOneStepAhead` and `waysTwoStepsAhead`) to represent the results of the last two states. This results in an **O(1)** space complexity.


### Time and Space Complexity

- **Time Complexity**: `O(n)` - We iterate through the string once.
- **Space Complexity**: `O(1)` - No additional data structures are used; only two variables (`waysOneStepAhead` and `waysTwoStepsAhead`) are maintained.


### Run

   ```bash
   cd logical
   go run main.go
   ```

---

# Blog API

## **Setup and Run**

**If you using same terminal with logical assessment**:

 ```bash
   cd ..
   ```

**Run Locally**:

```bash
   make run
   ```

**Run in Docker**:
   ```bash
   make docker-build
   make docker-run
   ```

**Run Linter**:
   ```bash
   make lint
   ```

---

# Endpoints

### **1. Get All Posts**

#### Success Scenario:
```bash
curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts?page=1&limit=2"
```

#### Get by ID:
```bash
curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1"
```

#### Edge Cases:
- Invalid Query Parameters will return first 10:
  ```bash
  curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts?page=abc&limit=-1"
  ```

---

### **2. Get Post by ID**

#### Success Scenario:
```bash
curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1"
```

#### Edge Cases:
- Non-existent ID:
  ```bash
  curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/9999"
  ```

- Invalid ID:
  ```bash
  curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/abc"
  ```

---

### **3. Create Post**

#### Success Scenario:
```bash
curl -X POST "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts" \
-H "Content-Type: application/json" \
-d '{"title":"New Post","content":"This is the content","author":"AuthorName"}'
```

#### Edge Cases:
- Missing Title:
  ```bash
  curl -X POST "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts" \
  -H "Content-Type: application/json" \
  -d '{"content":"This is content without title","author":"AuthorName"}'
  ```

- Invalid JSON:
  ```bash
  curl -X POST "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts" \
  -H "Content-Type: application/json" \
  -d '{"title":"Unclosed JSON"'
  ```

---

### **4. Update Post**

#### Success Scenario:
```bash
curl -X PUT "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1" \
-H "Content-Type: application/json" \
-d '{"title":"Updated Title","content":"Updated content","author":"Updated Author"}'
```

#### Edge Cases:
- Non-existent ID:
  ```bash
  curl -X PUT "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/9999" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title"}'
  ```

- Invalid ID:
  ```bash
  curl -X PUT "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/abc" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title"}'
  ```

- Invalid JSON:
  ```bash
  curl -X PUT "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1" \
  -H "Content-Type: application/json" \
  -d '{"title":'
  ```

---

### **5. Patch Post**

#### Success Scenario:
```bash
curl -X PATCH "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1" \
-H "Content-Type: application/json" \
-d '{"title":"Partially Updated Title"}'
```

#### Edge Cases:
- Empty Body:
  ```bash
  curl -X PATCH "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1" \
  -H "Content-Type: application/json" \
  -d '{}'
  ```

- Invalid JSON:
  ```bash
  curl -X PATCH "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1" \
  -H "Content-Type: application/json" \
  -d '{"title":'
  ```

---

### **6. Delete Post**

#### Success Scenario:
```bash
curl -X DELETE "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/1"
```

#### Edge Cases:
- Non-existent ID:
  ```bash
  curl -X DELETE "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/9999"
  ```

- Invalid ID:
  ```bash
  curl -X DELETE "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/abc"
  ```

---

## **Testing**

### **Run All Tests**
```bash
make test
```

### **Run Unit Tests**
```bash
make unit-test
```

### **Run Integration Tests**
```bash
make integration-test
```

### **Run Fuzz Tests**
```bash
make fuzz-test
```

### **Run Linter**
```bash
make lint
```

---

## **Graceful Shutdown**

The server supports graceful shutdown. To test:
1. Start the server:
   ```bash
   make run
   ```

2. Send a termination signal:
   ```bash
   kill -SIGTERM <pid>
   ```

The server will clean up resources and shut down gracefully.

---