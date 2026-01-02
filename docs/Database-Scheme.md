%%{init: {
  "theme": "default",
  "themeCSS": [
    ".er.relationshipLabel { fill: black; }", 
    ".er.relationshipLabelBox { fill: white; }", 
    ".er.entityBox { fill: lightgray; }",
    "[id^=entity-some] .er.entityBox { fill: lightgreen;} ",
    "[id^=entity-mytable] .er.entityBox { fill: powderblue;} ",
    "[id^=entity-anothertable] .er.entityBox { fill: pink;} "
    ]
}}%%
erDiagram
    users ||--o{ organization_founders : "user_id"
    users ||--o{ organization_members : "user_id"
    
    organizations ||--o{ organization_founders : "organization_id"
    organizations ||--o{ organization_members : "organization_id "
    organizations ||--o{ departments : "organization_id"

    organization_members ||--o{ employees : "member_id"
    
    departments ||--o{ positions : "department_id"
    positions ||--o{ employees : "position_id"
    
    departments ||--o{ departments : "parent"

    users {
        bigint id PK
        string email UK
        string phone UK
        string password_hash
        string first_name
        string last_name
        string middle_name
        boolean is_email_verified
        boolean is_phone_verified
        timestamptz created_at
    }
    
    organizations {
        bigint id PK
        string name
        string legal_name
        string inn
        string ogrn
        string kpp
        string legal_address
        org_status status "ENUM: draft,pending,approved,rejected"
        timestamptz created_at
    }
    
    organization_founders {
        bigint id PK
        bigint organization_id FK
        bigint user_id FK
        numeric share_percent
        boolean is_main
    }
    
    organization_members {
        bigint id PK
        bigint organization_id FK
        bigint user_id FK
        member_status status "ENUM: invited,active,blocked"
        timestamptz joined_at
    }
    
    
    departments {
        bigint id PK
        bigint organization_id FK
        bigint parent_id FK "self-reference"
        string name
        string description
    }
    
    positions {
        bigint id PK
        bigint organization_id FK
        bigint department_id FK
        string name
        boolean is_admin
        string description
    }
    
    employees {
        bigint id PK
        bigint member_id FK
        bigint position_id FK
        boolean is_intern
        date start_date
        date end_date
    }