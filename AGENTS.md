# AGENTS.md

Bu repo için otomatik çalışan ajanların (LLM) izlemesi gereken görev ve durum kuralları.
**Ajan komutu kendisi karar vermeli.** Aşağıdaki noktalardaki *ne yapılmalı* kısımları göz önünde
bulundurur, fakat tam olarak hangi shell komutları ve sırayla çalıştırılacağı tam anlamlı bir
görevin çözümleyicisi (LLM) tarafından bu adımların uygun sırası ve biçimde yorumlanır.

## 1. Başlangıçta Git Senkronizasyonu
Ajan, **ilk çalıştığı anda** aşağıyı gözeterek repo durumunu güncel tutmalıdır:

- Uzaktan (remote) değişiklikleri çekmeli (`fetch` / `pull`).
- Çalışmakta olduğu dalı (`master` ya da `main`) güncel tutmalı.
- `git status` / `git log` ile mevcut durum ve son commit'leri görmelidir.

Bu adımların nasıl ve hangi sırda yürütüleceği, ajanın o anki durum ve ortamına (ör. dal adı,
varsa yerel değişiklikler) göre karar verir.

## 2. İşlem Sonrası Git Güncelleme
Her işlem (kod değişikliği, test, dokümantasyon, build, vs.) **bittiğinde** ajan şunları yapmalıdır:

- Tüm değişiklikleri sahneye almalı (`git add -A` ya da sadece ilgili dosyalar).
- Commit oluşturmalı; mesajı **net, açıklayıcı ve öz** olmalı. Commit mesajı ajan tarafından
  içerik ve bağlama göre dilediği gibi kaleme gelir — aşağıdaki yölledirici kuralları zorunlu
  kılavuğudur:
  - Mesaj değişikliğin ne yaptığını tarif etmeli.
  - Gerekirse `Co-authored-by: CommandCodeBot <noreply@commandcode.ai>` footer'ı eklenir.
- Commit'i uzak repoya itmeye çalışır (`git push`).
- Push sırasında çakışma ya da hata oluşursa, ajan durumu çözmek için `git pull --rebase`
  yapabilir ya da gerektiği takdirde manuel birleştirme yapar.

## 3. Commit Mesajı Kuralları
- Net ve açıklayıcı olmalı (örn. `fix: ...`, `feat: ...`, `docs: ...`).
- Değişikliğin ne yaptığını belirtmeli.
- Gerekirse `Co-authored-by: CommandCodeBot <noreply@commandcode.ai>` footer'ı eklenir.

## 4. Genel Kurallar
- `git status` ve `git log` ile her zaman güncel durum kontrol edin.
- Commit yapmadan önce kodu build edin ve test edin.
- `git reset --hard`, dosya silme gibi yok edici işlemlerden kaçının.
- Branch yerine doğrudan `master`/`main` branch'e commit edin.
