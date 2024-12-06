
# Blog API - AWS integration

### **1. Get All Posts**

#### Success Scenario:
```bash
curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts?page=1&limit=2"
```

#### Get by ID:
```bash
curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/d6a03097-347f-4653-b10e-ee195849805a"
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
curl -X GET "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/d6a03097-347f-4653-b10e-ee195849805a"
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
curl -X PUT "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/d6a03097-347f-4653-b10e-ee195849805a" \
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
curl -X PATCH "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/d6a03097-347f-4653-b10e-ee195849805a" \
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
curl -X DELETE "https://80fapksm9d.execute-api.us-east-1.amazonaws.com/v1/posts/8d04d6ac-51b1-48ad-8a26-109734027132"
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
