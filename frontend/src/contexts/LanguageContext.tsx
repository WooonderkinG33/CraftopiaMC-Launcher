import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';

export type Lang = 'ru' | 'en';

export const dict = {
  ru: {
    sidebarPlay: "Играть",
    sidebarSettings: "Настройки",
    sidebarSupport: "Поддержка",
    sidebarExit: "Выход",
    statusReady: "ГОТОВ К ЗАПУСКУ",
    statusRunning: "ИГРА УСПЕШНО ЗАПУЩЕНА",
    stepSync: "Синхронизация...",
    stepJava: "Загрузка Java",
    stepInstall: "Установка...",
    stepMcClient: "Загрузка Minecraft client",
    stepMcLibs: "Загрузка Minecraft libs",
    stepMcAssets: "Загрузка Minecraft assets",
    stepFabric: "Загрузка Fabric",
    stepBuildClient: "Создание игрового клиента",
    stepLaunch: "Запуск игры...",
    errConnection: "Ошибка соединения: перезапуск через",
    settingsTitle: "НАСТРОЙКИ ЛАУНЧЕРА",
    ramAllocation: "Выделение ОЗУ",
    languageLabel: "Язык интерфейса",
    reinstallLauncher: "СБРОС ФАЙЛОВ ЛАУНЧЕРА",
    reinstallConfirm: "Вы уверены? Это приведет к полной очистке всех временных файлов игры.",
    reinstalledSuccess: "Файлы лаунчера успешно удалены!",
    close: "Закрыть",
    ramLimit: "Лимит оперативной памяти",
    telegram: "Наш Телеграм",
    support: "Техподдержка",
    version: "Версия лаунчера",
    // Phase status translations
    init: "ИНИЦИАЛИЗАЦИЯ...",
    javaChecking: "Java: проверка",
    javaDownloading: "Java: загрузка",
    javaUnpacking: "Java: распаковка",
    javaOk: "Java: OK",
    mcFetch: "Minecraft: получение манифеста",
    mcClientCheck: "Minecraft: проверка клиента",
    mcClientDownloading: "Minecraft: загрузка клиента",
    mcClientOk: "Minecraft: клиент OK",
    mcLibsCheck: "Minecraft: проверка библиотек",
    mcLibsDownloading: "Minecraft: загрузка библиотек",
    mcLibsOk: "Minecraft: библиотеки OK",
    mcAssetsCheck: "Minecraft: проверка ассетов",
    mcAssetsDownloading: "Minecraft: загрузка ассетов",
    mcAssetsOk: "Minecraft: ассеты OK",
    fabricDownloading: "Fabric: загрузка",
    fabricOk: "Fabric: OK",
    launching: "Запуск игры...",
    gameRunning: "ИГРА ЗАПУЩЕНА",
    errorLaunch: "Ошибка запуска, проверьте логи",
  },
  en: {
    sidebarPlay: "Play",
    sidebarSettings: "Settings",
    sidebarSupport: "Support",
    sidebarExit: "Exit",
    statusReady: "READY TO LAUNCH",
    statusRunning: "GAME SUCCESSFULLY LAUNCHED",
    stepSync: "Synchronization...",
    stepJava: "Downloading Java",
    stepInstall: "Installing...",
    stepMcClient: "Downloading Minecraft client",
    stepMcLibs: "Downloading Minecraft libs",
    stepMcAssets: "Downloading Minecraft assets",
    stepFabric: "Downloading Fabric",
    stepBuildClient: "Creating game client",
    stepLaunch: "Launching game...",
    errConnection: "Connection error: restarting in",
    settingsTitle: "LAUNCHER SETTINGS",
    ramAllocation: "RAM Allocation",
    languageLabel: "Interface Language",
    reinstallLauncher: "RESET LAUNCHER FILES",
    reinstallConfirm: "Are you sure? This will completely clear all temporary game files.",
    reinstalledSuccess: "Launcher files successfully deleted!",
    close: "Close",
    ramLimit: "RAM Limit",
    telegram: "Our Telegram",
    support: "Help & Support",
    version: "Launcher Version",
    // Phase status translations
    init: "INITIALIZING...",
    javaChecking: "Java: checking",
    javaDownloading: "Java: downloading",
    javaUnpacking: "Java: unpacking",
    javaOk: "Java: OK",
    mcFetch: "Minecraft: fetching manifest",
    mcClientCheck: "Minecraft: checking client",
    mcClientDownloading: "Minecraft: downloading client",
    mcClientOk: "Minecraft: client OK",
    mcLibsCheck: "Minecraft: checking libraries",
    mcLibsDownloading: "Minecraft: downloading libraries",
    mcLibsOk: "Minecraft: libraries OK",
    mcAssetsCheck: "Minecraft: checking assets",
    mcAssetsDownloading: "Minecraft: downloading assets",
    mcAssetsOk: "Minecraft: assets OK",
    fabricDownloading: "Fabric: downloading",
    fabricOk: "Fabric: OK",
    launching: "Launching game...",
    gameRunning: "GAME RUNNING",
    errorLaunch: "Launch error, check the logs",
  }
};

interface LanguageContextType {
  lang: Lang;
  setLang: (lang: Lang) => void;
  t: typeof dict['ru'];
  tPhase: (key: string, pct?: number) => string;
}

const ctx = createContext<LanguageContextType>({ lang: 'ru', setLang: () => {}, t: dict.ru, tPhase: () => '' });

export const LanguageProvider = ({ children, initialLang }: { children: ReactNode; initialLang?: Lang }) => {
  const [lang, setLang] = useState<Lang>(initialLang || 'ru');
  const t = dict[lang];
  const tPhase = useCallback((key: string, pct?: number) => {
    const phrase = (t as any)[key] || key;
    return pct !== undefined ? `${phrase} ${pct}%` : phrase;
  }, [t]);
  return (
    <ctx.Provider value={{ lang, setLang, t, tPhase }}>
      {children}
    </ctx.Provider>
  );
};

export const useLanguage = () => useContext(ctx);
