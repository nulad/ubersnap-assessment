## **Technical Assessment: Scalable Coupon System**

### **Objective**
Build a REST API in **Golang** for a "Flash Sale" Coupon System. The system must handle high concurrency, guarantee strict data consistency, and be easy to deploy via Docker.

**IMPORTANT:** We will use an **automated script** to test your submission. You must follow the API Specifications exactly as described below.

### **Tech Stack Requirements**
* **Language:** Golang.
* **Database:** MongoDB or SQL (PostgreSQL/MySQL).
* **Infrastructure:** Application must run via **Docker / Docker Compose**.

### **Database Constraints & Rules**
1. **Separation of Concerns:** You must separate Coupon data and Claim history into two distinct Collections/Tables.
2. **No Embedding:** You are **not allowed** to embed claim history inside the Coupon record.
3. **Uniqueness Rule (CRITICAL):**
    * A `user_id` can claim a specific `coupon_name` **only once**.
    * However, the same `user_id` **can** claim other different coupons.
    * *Hint:* Ensure your database schema enforces uniqueness on the pair `(user_id, coupon_name)` to prevent race conditions.

---

### **API Specifications (Strict Contract)**

You must implement the following routes with the exact payload structure.

#### **1. Create Coupon**
Registers a new coupon into the system.

* **Endpoint:** `POST /api/coupons`
* **Request Body:** `{"name": "PROMO_SUPER", "amount": 100}`
* **Response:** `201 Created`

#### **2. Claim Coupon**
Attempts to claim a coupon for a specific user.

* **Endpoint:** `POST /api/coupons/claim`
* **Header:** `Content-Type: application/json`
* **Request Body:**
    ```json
    {
      "user_id": "user_12345",
      "coupon_name": "PROMO_SUPER"
    }
    ```
* **Logic & Behavior:**
    * **Check Eligibility:** If `user_12345` has already claimed `PROMO_SUPER`, reject immediately.
    * **Check Stock:** If stock is 0, reject.
    * **Concurrency Safety:** The process of checking stock, inserting claim, and deducting stock must be **Atomic** (using Database Transactions).
* **Response Codes:**
    * **Success:** `200` or `201`.
    * **Rejected (Already Claimed):** `409 Conflict` (Preferred) or `400 Bad Request`.
    * **Rejected (No Stock):** `400` to `409`.

#### **3. Get Coupon Details**

* **Endpoint:** `GET /api/coupons/{name}`
* **Response Body:**
    ```json
    {
      "name": "PROMO_SUPER",
      "amount": 100,
      "remaining_amount": 0,
      "claimed_by": ["user_12345", ...]
    }
    ```

---

### **Evaluation Scenarios**

Your code will be stress-tested with the following scenarios:

1. **The "Flash Sale" Attack:** 50 concurrent requests for a coupon with only 5 items in stock. (Result must be exactly 5 claims, 0 remaining).
2. **The "Double Dip" Attack:** 10 concurrent requests from the **SAME** user for the same coupon. (Result must be exactly 1 success, 9 failures).

---

### **Submission Guidelines**

Once you have completed the challenge, please follow these steps to submit your application:

#### **1. GitHub Repository**
* Upload your full source code to a **GitHub Repository**.
* **Repository Access:** Ideally, make it **Public** to ensure we can clone it immediately. If it must be Private, please be ready to grant access upon request.

#### **2. README.md Requirements**
Your repository **MUST** include a `README.md` file containing:
* **Prerequisites:** What needs to be installed (e.g., Docker Desktop).
* **How to Run:** The exact command to start the application (e.g., `docker-compose up --build`).
* **How to Test:** Instructions on how to run your tests or how we can trigger the endpoints.
* **Architecture Notes:** A brief explanation of your database design and locking strategy.

#### **3. Email Submission**
Send an email to verify your submission.

* **To:** `rofie@ubersnap.com`
* **Cc:** `boonchin@ubersnap.com`
* **Subject:** Backend Engineer Test - [Your Name]
* **Body:**
    * Link to your GitHub Repository.
    * Any specific notes you want to add.
* **Attachment:** Please attach your latest **CV / Resume** (PDF format).

**We look forward to reviewing your code!**