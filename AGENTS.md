# AGENTS.md

Repo için otomatik çalışan ajanlar için yönergeler.

## Başlangıçta Git Senkronizasyonu
Her ajan, **ilk çalıştığında** şu adımları izler:

```bash
git fetch origin
git checkout main   # ya da master
git pull origin main
git status
```

Bu sayede ajan, çalışmaya başlamadan önce en güncel kodu alır ve mevcut değişiklikleri görür.

## İşlem Sonrası Git Güncelleme
Her işlem (kod değişikliği, test, dokümantasyon vs.) **bittiğinde**, ajan şu adımları izler:

```bash
git add -A
git commit -m "<açıklayıcı commit mesajı>"
git push origin main
```

Commit mesajı kuralları:
- Net ve açıklayıcı olmalı.
- Değişikliğin ne yaptığını belirtmeli.
- Gerekirse aşağıdaki co-author footer'ı eklenir:

```
Co-authored-by: CommandCodeBot <noreply@commandcode.ai>
```

## Genel Kurallar
- Git durumunu her zaman kontrol edin (`git status`).
- Commit yapmadan önce kodu build edin ve test edin.
- `git reset --hard` gibi yok edici işlemlerden kaçının.
- Branch yerine doğrudan `main`/`master` branch'e commit edin.
