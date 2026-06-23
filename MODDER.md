# Fabric Mod — кастомный тайтл и иконка окна Minecraft

Лаунчер передаёт два Java system property при запуске Minecraft.
Мод должен их прочитать и применить.

## System properties

| Property | Тип | Описание | Пример |
|----------|-----|----------|--------|
| `craftopiamc.iconPath` | `String` | Полный путь до PNG иконки лаунчера | `~/.craftopiamc/runtime/launcher-icon.png` |
| `craftopiamc.launcherVersion` | `String` | Версия лаунчера | `2026.6.24.0050` |

## Где вставить код

Создай класс, который имплементирует `ModInitializer` и в методе `onInitialize()` читает свойства и применяет их.

### Чтение свойств

```java
String iconPath = System.getProperty("craftopiamc.iconPath");
String launcherVersion = System.getProperty("craftopiamc.launcherVersion");
```

### Смена тайтла окна

Тайтл нужно менять ПОСЛЕ того как Minecraft создаст окно (после инициализации LWJGL).
Лучше всего через Mixin в `MinecraftClient` или через `ClientTickEvents.START` (первый тик).

```java
ClientTickEvents.START.register(client -> {
    String version = System.getProperty("craftopiamc.launcherVersion");
    if (version != null && client.getWindow() != null) {
        client.getWindow().setTitle("CraftopiaMC Launcher " + version);
    }
});
```

### Смена иконки окна

Иконку можно загрузить и установить один раз при старте:

```java
String iconPath = System.getProperty("craftopiamc.iconPath");
if (iconPath != null) {
    try {
        File iconFile = new File(iconPath);
        if (iconFile.exists()) {
            // Загружаем PNG и устанавливаем как иконку окна
            InputStream is = Files.newInputStream(iconFile.toPath());
            // Используй GLFW для установки иконки
            // Или через MinecraftClient.getInstance().getWindow().setIcon(...)
        }
    } catch (IOException e) {
        // log error
    }
}
```

### Через GLFW напрямую

```java
import org.lwjgl.glfw.GLFW;
import org.lwjgl.system.MemoryStack;
import java.nio.ByteBuffer;
import javax.imageio.ImageIO;
import java.awt.image.BufferedImage;
import java.io.File;
import javax.imageio.ImageIO;

// В начале игры (например, в onInitializeClient):
long window = GLFW.glfwGetCurrentContext();
if (window != 0L) {
    String iconPath = System.getProperty("craftopiamc.iconPath");
    if (iconPath != null) {
        BufferedImage img = ImageIO.read(new File(iconPath));
        // Конвертировать BufferedImage в GLFWImage и установить
        // через GLFW.glfwSetWindowIcon(window, ...)
    }
    
    // Сменить тайтл
    String version = System.getProperty("craftopiamc.launcherVersion");
    if (version != null) {
        GLFW.glfwSetWindowTitle(window, "CraftopiaMC Launcher " + version);
    }
}
```

## Структура проекта Fabric

```
src/main/java/com/craftopiamc/launcherbrand/
├── CraftopiaBrand.java        // implements ModInitializer или ClientModInitializer
└── resources/
    └── fabric.mod.json
```

### fabric.mod.json

```json
{
  "schemaVersion": 1,
  "id": "craftopiabrand",
  "name": "CraftopiaMC Brand",
  "version": "1.0.0",
  "environment": "client",
  "entrypoints": {
    "client": ["com.craftopiamc.launcherbrand.CraftopiaBrand"]
  },
  "depends": {
    "fabricloader": ">=0.16.0",
    "minecraft": ">=26.1.2"
  }
}
```

## Важно

- Мод должен быть **client-only** (не ставится на сервер)
- Тайтл менять **после** инициализации окна (не в конструкторе)
- Иконку загружать однократно, не каждый тик
- Если свойства `null` — ничего не делать (лаунчер без брендинга)
