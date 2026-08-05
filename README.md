# Hospital Middleware API

This is a middleware API for hospital management system. It provides endpoints for managing hospitals, patients, and staff.

# ชื่อโปรเจกต์
    hospital-middleware-api
    ระบบกลางสำหรับโรงพยาบาล (Hospital Middleware API) สำหรับการจัดการข้อมูลโรงพยาบาล, ผู้ป่วย, และบุคลากรทางการแพทย์

## Tech Stack
  Go 1.25 / Gin / PostgreSQL 16 / nginx / Docker Compose

## Getting Started
  ต้องมีอะไรติดตั้งบ้าง → คำสั่งรัน → วิธีเช็คว่าขึ้นแล้ว
  ⚠️ ใส่ทั้ง `make` และ `docker compose` ตรงๆ (เครื่องคุณยังไม่มี make กรรมการก็อาจไม่มี)

## Project Structure
  ผังโฟลเดอร์ + อธิบายว่าแต่ละชั้นรับผิดชอบอะไร
  บอกด้วยว่าทำไมต้องมี internal/ (compiler บังคับ ไม่ใช่แค่ convention)

## Database Schema
  ER diagram + ตารางอธิบายแต่ละ table
  จุดขาย: อธิบายว่าทำไม unique ต้องผูกกับ hospital_id — นี่คือการตัดสินใจที่ควรอวด

## API Spec
  ทุก endpoint: method, path, request body, response, status code
  ตัวอย่าง curl ที่ก็อปไปรันได้จริง
  ตารางรหัสโรงพยาบาลที่ seed ไว้

## Design Decisions
  ⭐ ส่วนที่แยกคนได้คะแนนดีออกจากคนทั่วไป
  ทางเลือกที่คุณเจอ → เลือกอะไร → เพราะอะไร
  เช่น: ทำไมใช้ hospital_code ไม่ใช่ชื่อ / ทำไมไม่มี soft delete /
       ข้อมูลจาก HIS เข้าระบบตอนไหน

## Testing
  วิธีรัน test + ครอบคลุมอะไรบ้าง
