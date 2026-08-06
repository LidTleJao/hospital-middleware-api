# Hospital Middleware API

  hospital-middleware-api
  ระบบกลางสำหรับโรงพยาบาล (Hospital Middleware API) สำหรับการจัดการข้อมูลโรงพยาบาล, ผู้ป่วย, และบุคลากรทางการแพทย์

## Tech Stack
  Go 1.25 / Gin / PostgreSQL 16 / nginx / Docker Compose

## Getting Started
  ```bash
  cp .env.example .env
  docker compose up --build -d
  curl -i http://localhost:8080/health      # {"status":"ok"}
  docker compose logs migrate               # 1/u init
  go test ./...                             # 16 เคส
  ```

  ถ้าเครื่องมี PostgreSQL รันอยู่แล้วที่ port 5432 จะชนกับ container
  แก้ได้โดยเปลี่ยนค่าใน .env แล้ว `docker compose up -d` ใหม่:

      POSTGRES_PORT=55432

  แล้วต่อเครื่องมือดู DB ที่ localhost:55432 (user/password/db = hospital ทั้งหมด)


## Project Structure

  ```
  hospital-middleware-api/
  ├── cmd/                          # main entry point ของโปรเจกต์
  │   └── api/                      # main.go ของ API\
  ├── deploy/
  |   └── nginx/                    # nginx config
  ├── internal/                     # โค้ดที่ไม่อยากให้คนอื่น import
  |   ├── config/                   # อ่าน env → struct Config, fail fast ตอน boot
  |   ├── database/                 # เปิด connection pool + ping
  |   ├── hisclient/                # HTTP client ยิงไป HIS ภายนอก
  |   ├── model/                    # struct ที่ map กับตาราง (db + json tag)
  |   ├── repository/               # SQL ทั้งหมด
  |   ├── service/                  # business logic + bcrypt + JWT
  |   └── transport/http/
  |       ├── router.go             # ประกาศ route
  |       ├── handler/              # แปลง HTTP ↔ service
  |       └── middleware/           # ตรวจ token
  ├── migrations/                   # database migration files
  ```

## Database Schema

  ```mermaid
  erDiagram
      hospitals ||--o{ staffs   : "employs"
      hospitals ||--o{ patients : "has"

      hospitals {
          bigserial   hospital_id      PK
          varchar     hospital_code    UK "H001, H002, ..."
          varchar     hospital_name_th NULL
          varchar     hospital_name_en NULL
          timestamptz created_at
          timestamptz updated_at
      }

      staffs {
          bigserial   staff_id    PK
          bigint      hospital_id FK
          varchar     username         "unique คู่กับ hospital_id"
          varchar     password         "bcrypt hash"
          timestamptz created_at
          timestamptz updated_at
      }

      patients {
          bigserial   patient_id     PK
          bigint      hospital_id    FK
          varchar     patient_hn          "unique คู่กับ hospital_id"
          varchar     national_id    NULL "unique คู่กับ hospital_id"
          varchar     passport_id    NULL "unique คู่กับ hospital_id"
          varchar     first_name_th  NULL
          varchar     middle_name_th NULL
          varchar     last_name_th   NULL
          varchar     first_name_en  NULL
          varchar     middle_name_en NULL
          varchar     last_name_en   NULL
          date        date_of_birth
          varchar     phone_number   NULL
          varchar     email          NULL
          varchar     gender              "M หรือ F"
          timestamptz created_at
          timestamptz updated_at
      }
  ```

  ความสัมพันธ์ทั้งหมดวิ่งออกจาก `hospitals` — ทุกแถวใน `staffs` และ `patients` ผูกกับโรงพยาบาลเดียวเสมอ และ unique constraint ทุกตัวมี `hospital_id` นำหน้า จึงเป็นขอบเขตข้อมูลที่บังคับตั้งแต่ระดับฐานข้อมูล ไม่ใช่แค่ในโค้ด


  ตาราง hospital
  สาเหตุเพราะ: hospital_code เป็น unique identifier ของโรงพยาบาล ใช้ในการเชื่อมโยงข้อมูลระหว่างระบบต่าง ๆ และป้องกันความสับสนจากชื่อโรงพยาบาลที่อาจซ้ำกัน

  ตาราง hospital ที่ seed ไว้:
  | hospital_code | hospital_name_th | hospital_name_en |
  |---------------|------------------|------------------|
  | H001          | โรงพยาบาล A      | Hospital A       |
  | H002          | โรงพยาบาล B      | Hospital B       |
  | H003          | โรงพยาบาล C      | Hospital C       |
  | H004          | โรงพยาบาล D      | Hospital D       |
  | H005          | โรงพยาบาล E      | Hospital E       |

  ตาราง patient
  unique เป็นคู่กับ hospital_id ทั้ง 3 ตัว:(hospital_id, patient_hn) · (hospital_id, national_id) · (hospital_id, passport_id)
  เหตุผล: คนไข้คนเดียวกันไปรักษาได้หลายโรงพยาบาล ถ้า unique ที่ national_id เดี่ยวๆ ระบบจะinsert ไม่ได้ทันทีที่คนไข้ไปโรงพยาบาลที่สอง

  CHECK (national_id IS NOT NULL OR passport_id IS NOT NULL) เหตุผล: คนต่างชาติมีแต่พาสปอร์ตคนไทยมีแต่บัตรประชาชน แต่ถ้าไม่มีทั้งคู่จะเป็นแถวที่ค้นไม่มีวันเจอ

  ตาราง staff
  unique (hospital_id, username) — ชื่อผู้ใช้ซ้ำข้ามโรงพยาบาลได้จึงเป็นเหตุผลที่ /staff/login ต้องรับ hospital มาด้วย

## API Spec

  ฟังก์ชันการทำงานหลักของ API:
  staff
  - POST /staff/create: สร้างบัญชีบุคลากรทางการแพทย์ใหม่

  สร้างบัญชีเจ้าหน้าที่ ไม่ต้อง login

  ** Requset **
  ```json
  {
    "username":"doc_x",
    "password":"secret123",
    "hospital":"H001"
  }
  ```

  ** Response 201**
  ```json
  {
    "id":40,
    "username":"doc_x",
    "hospital":"H001"
  }
  ```

  ** Error **
  | Status Code | Error Message         |
  |-------------|-----------------------|
  | 400         | Invalid request body  |
  | 404         | Hospital not found    |
  | 500         | Internal server error |

  - POST /staff/login: เข้าสู่ระบบบุคลากรทางการแพทย์

  ล็อกอินเจ้าหน้าที่ → ได้ token เอาไปใช้กับ /patient/import และ /patient/search

  ** Request **
  ```json
  {
    "username":"doc_x",
    "password":"secret123",
    "hospital":"H001"
  }
  ```

  ** Response 200**
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```

  ** Error **
  | Status Code | Error Message         |
  |-------------|-----------------------|
  | 400         | Invalid request body  |
  | 401         | Invalid credentials   |
  | 500         | Internal server error |

  patient
  - POST /patient/import: นำเข้าข้อมูลผู้ป่วยจากระบบ HIS

  นำเข้าข้อมูลผู้ป่วยจากระบบ HIS → ต้อง login staff ก่อน
  ต้องมี Authorization header: Bearer <token> จาก /staff/login

  ** Request **
  ```json
  {
    "id": "P001"
  }
  ```

  ** Response 200**
  ```json
  {
    "patient_id": "P001",
    "hospital_id": "H001",
    "first_name_th": "สมชาย",
    "middle_name_th": "ใจดี",
    "last_name_th": "ใจดี",
    "first_name_en": "Somchai",
    "middle_name_en": "Jaidee",
    "last_name_en": "Jaidee",
    "date_of_birth": "1990-01-01",
    "patient_hn": "123456",
    "national_id": "1234567890123",
    "passport_id": "A1234567",
    "phone_number": "0812345678",
    "email": "example@gmail.com",
    "gender": "M"
  }
  ```

  ** Error **
  | Status Code | Error Message            |
  |-------------|--------------------------|
  | 400         | Invalid request body     |
  | 401         | Unauthorized             |
  | 404         | Patient not found in HIS |
  | 500         | Internal server error    |


  - POST /patient/search: ค้นหาผู้ป่วยตามเงื่อนไขที่กำหนด

  สำหรับค้นหาผู้ป่วย → ต้อง login staff ก่อน
  ต้องมี Authorization header: Bearer <token> จาก /staff/login
  เมื่อไม่พบผู้ป่วยที่ตรงเงื่อนไข จะได้ 200 พร้อม []  (ไม่ใช่ null และไม่ใช่ 404)

  ** Request **
  ```json
  {
    "first_name": "สมชาย",
    "middle_name": "ใจดี",
    "last_name": "ใจดี",
    "date_of_birth": "1990-01-01",
    "national_id": "1234567890123",
    "passport_id": "A1234567",
    "phone_number": "0812345678",
    "email": "example@gmail.com",
    "page": 1,
    "page_size": 10
  }
  ```

  ** Response 200**
  ```json
  {
    {
      "patient_id": "P001",
      "hospital_id": "1",
      "first_name_th": "สมชาย",
      "middle_name_th": "ใจดี",
      "last_name_th": "ใจดี",
      "first_name_en": "Somchai",
      "middle_name_en": "Jaidee",
      "last_name_en": "Jaidee",
      "date_of_birth": "1990-01-01",
      "patient_hn": "123456",
      "national_id": "1234567890123",
      "passport_id": "A1234567",
      "phone_number": "0812345678",
      "email": "example@gmail.com",
      "gender": "M"
    }
  }
  ```

  ** Response 200 กรณีไม่เจอ**
  ```json
  {
    []
  }
  ```

  ** Error **
  | Status Code | Error Message         |
  |-------------|-----------------------|
  | 400         | Invalid request body  |
  | 401         | Unauthorized          |
  | 500         | Internal server error |

## Design Decisions

  1. hospital รับเป็นรหัส (H001) ไม่ใช่ชื่อ — ชื่อเปลี่ยนได้/มีปัญหาตัวพิมพ์-ช่องว่าง-2 ภาษาและรหัสไม่ใช่มาตรการความปลอดภัยความปลอดภัยจริงอยู่ที่ token
  2. unique ทุกตัวผูกกับ hospital_id — คนไข้คนเดียวไปได้หลายโรงพยาบาล, staffชื่อซ้ำข้ามโรงพยาบาลได้
  3. hospital_id มาจาก token ไม่ใช่ request body — /patient/search จึงไม่มี field hospital เลย
  4. login ตอบ 401 เหมือนกันทุกกรณี แต่ create ตอบ 404 ได้ — กัน user enumeration
  5. ไม่มี soft delete — โจทย์ไม่มี API ลบ และมันจะบังคับให้ unique ทุกตัวเป็น partial index
  6. HIS ยิงไม่ติดจริง — เขียน client ตามสเปกจริง + มี /patient/import ให้ดึงเข้ามา แต่ทดสอบด้วย stub เพราะ hospital-a.api.co.th ไม่มีอยู่จริง

## Testing

  ```bash

    go test ./...              # รันทั้งหมด
    go test ./... -v           # ดูรายชื่อเคสทั้งหมด
    go test ./... -cover       # ดู coverage

  ```

  | endpoint | positive | negative |
  |---|---|---|
  | POST /staff/create | สร้างสำเร็จ | ไม่มีโรงพยาบาล · password สั้น · ไม่ส่ง hospital · error ภายใน |
  | POST /staff/login | ได้ token | credential ผิด · โรงพยาบาลผิดแยกไม่ออกจาก password ผิด · error ภายใน |
  | POST /patient/search | คืนรายการผู้ป่วย | body ผิดรูป · error ภายใน |
  | POST /patient/import | นำเข้าสำเร็จ | ไม่พบใน HIS · ไม่ส่ง id |

  เทสยังล็อกพฤติกรรมด้านความปลอดภัยไว้ด้วย:

  - ทุก response ถูกตรวจว่าไม่มี field `password` หลุดออกไป
  - `unknown_hospital_is_indistinguishable_from_a_wrong_password` — ยืนยันว่า login ตอบเหมือนกันทุกกรณี ถ้าวันหนึ่งมีคนเผลอแยก error ให้ละเอียดขึ้นเทสจะพังทันที
  - ค้นไม่เจอต้องได้ `[]` ไม่ใช่ `null`
