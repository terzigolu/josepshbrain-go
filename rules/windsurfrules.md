# Ramorie MCP - Windsurf Rules

> Bu dosyayı `.windsurfrules` olarak projenize kopyalayın.

## MCP Server Yapılandırması

Ramorie MCP server'ı Windsurf'te kullanmak için `~/.windsurf/settings.json` veya proje `.windsurf/mcp.json` dosyasına ekleyin:

```json
{
  "mcpServers": {
    "ramorie": {
      "command": "ramorie",
      "args": ["mcp", "serve"],
      "env": {}
    }
  }
}
```

---

## Agent Kuralları

### 🎯 Temel Prensipler

1. **Her oturumda bağlamı kontrol et**
   - İlk iş: `get_active_context_pack` çağır
   - Aktif context yoksa: `list_context_packs` → `activate_context_pack`

2. **Görev odaklı çalış**
   - Yeni iş = Yeni task: `create_task` + `start_task`
   - İlerlemeyi kaydet: `add_task_note`
   - Bitince: `complete_task`

3. **Bilgiyi sakla**
   - Öğrenilenleri kaydet: `add_memory`
   - Mevcut bilgiyi ara: `recall`

4. **Kararları belgele**
   - Önemli teknik kararlar: `create_decision`
   - ADR formatı kullan (context, consequences)

---

## 📋 Görev Yönetimi Kuralları

### Yeni Görev Başlatma
```
1. get_active_context_pack    → Bağlamı kontrol et
2. create_task                → Görevi oluştur
   - description: "Net ve açıklayıcı başlık"
   - priority: H/M/L
3. start_task                 → Görevi başlat
4. add_task_note              → İlk planı kaydet
```

### Çalışma Sırasında
```
- Her anlamlı ilerleme → add_task_note
- Her 25% ilerleme → update_progress
- Öğrenilen bilgi → add_memory
- Alınan karar → create_decision
```

### Görev Tamamlama
```
1. add_task_note              → Son durumu kaydet
2. complete_task              → Görevi tamamla
3. add_memory                 → Öğrenilenleri sakla (opsiyonel)
```

---

## 🧠 Bilgi Yönetimi Kuralları

### Memory Kullanımı
- **Kaydet**: Tekrar kullanılabilir her bilgiyi `add_memory` ile sakla
- **Ara**: Soru sormadan önce `recall` ile mevcut bilgiyi kontrol et
- **Bağla**: İlgili görevlerle `create_memory_task_link` kullan

### Memory İçerik Formatı
```
✅ İyi:
"PostgreSQL connection pooling: max_connections=100,
pool_size=20. Performans için pgbouncer önerilir."

❌ Kötü:
"db ayarları"
```

---

## 📝 Karar Kayıt Kuralları

### Ne Zaman Karar Kaydı Oluştur?
- Mimari değişiklikler
- Teknoloji seçimleri
- API tasarım kararları
- Güvenlik politikaları
- Performans trade-off'ları

### Karar Formatı
```json
{
  "title": "JWT yerine Session-based Auth",
  "description": "Kullanıcı oturumları için session tabanlı auth",
  "area": "Architecture",
  "status": "approved",
  "context": "Mobile app desteği için stateless gerekli değil,
              server-side session yönetimi daha güvenli",
  "consequences": "Redis session store gerekecek,
                   horizontal scaling için sticky sessions"
}
```

---

## 🔄 Bağlam Yönetimi Kuralları

### Context Pack Kullanımı
- Her proje/özellik için ayrı context pack
- Konu değiştiğinde `activate_context_pack`
- Aktif context = Agent'ın odak noktası

### Context Değişikliği
```
1. stop_task                  → Mevcut görevi duraklat
2. activate_context_pack      → Yeni bağlamı aktifle
3. get_next_tasks             → Yeni görevleri al
4. start_task                 → Yeni göreve başla
```

---

## ⚡ Hızlı Referans

| İşlem | Tool |
|-------|------|
| Görev oluştur | `create_task` |
| Görevi başlat | `start_task` |
| Not ekle | `add_task_note` |
| İlerleme güncelle | `update_progress` |
| Görevi tamamla | `complete_task` |
| Bilgi sakla | `add_memory` |
| Bilgi ara | `recall` |
| Karar kaydet | `create_decision` |
| Bağlam değiştir | `activate_context_pack` |
| Aktif görev | `get_active_task` |

---

## 🚫 Yapılmaması Gerekenler

1. ❌ Görev oluşturmadan çalışmaya başlama
2. ❌ İlerlemeyi kaydetmeden uzun süre çalışma
3. ❌ Önemli kararları belgelemeden geçme
4. ❌ Memory'leri etiketsiz bırakma
5. ❌ Bağlamı kontrol etmeden yeni işe başlama

---

*Ramorie MCP v1.7.0 - 57 tool desteklenmektedir*
