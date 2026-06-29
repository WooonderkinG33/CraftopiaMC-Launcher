import { useState, useEffect } from 'react';
import DesktopLayout from './layouts/DesktopLayout';
import PlayPage from './pages/Play';
import { LauncherProvider } from './store/LauncherContext';
import { LanguageProvider } from './contexts/LanguageContext';
import './index.css';

declare global {
  interface Window {
    __launcherSettings?: { language: string; ram_mb: number; max_ram: number };
    go?: {
      main?: {
        App?: {
          GetSettings?: () => Promise<{ Language: string; RamMB: number; MaxRam: number }>;
          SaveSettings?: (language: string, ramMB: number) => Promise<boolean>;
          StartRuntimePreparation?: () => void;
          HideToTray?: () => void;
          KillMinecraft?: () => void;
          ResetLauncher?: () => Promise<boolean>;
        };
      };
    };
  }
}

const LS_KEY = 'craftopia_settings';

function writeLS(language: string, ramMB: number, maxRam: number) {
  try { localStorage.setItem(LS_KEY, JSON.stringify({ language, ramMB, maxRam })); } catch (_) {}
}

export default function App() {
  const [settings, setSettings] = useState<{ language: string; ramMB: number; maxRam: number } | null>(null);

  useEffect(() => {
    let cancelled = false;
    let attempts = 0;

    const tryLoad = async () => {
      // 1) Try window.__launcherSettings (WindowExecJS — fastest, but may race)
      const win = window.__launcherSettings;
      if (win && win.language) {
        if (!cancelled) {
          setSettings({ language: win.language, ramMB: win.ram_mb, maxRam: win.max_ram });
          writeLS(win.language, win.ram_mb, win.max_ram);
        }
        return;
      }

      // 2) Try localStorage (persists across normal restarts, lost on self-update)
      try {
        const raw = localStorage.getItem(LS_KEY);
        if (raw) {
          const p = JSON.parse(raw);
          if (p && p.language) {
            if (!cancelled) setSettings({ language: p.language, ramMB: p.ramMB, maxRam: p.maxRam || 16384 });
            // Continue to try Go binding to get fresher data
          }
        }
      } catch (_) {}

      // 3) Poll Go binding (works even after self-update — reads from settings.json)
      for (let i = 0; i < 30 && !cancelled; i++) {
        try {
          if (window.go?.main?.App?.GetSettings) {
            const s = await window.go.main.App.GetSettings();
            if (!cancelled && s && s.Language) {
              const lang = s.Language || 'en';
              const ram = s.RamMB || 2048;
              const max = s.MaxRam || 16384;
              setSettings({ language: lang, ramMB: ram, maxRam: max });
              writeLS(lang, ram, max);
              return;
            }
          }
        } catch (_) {}
        await new Promise(r => setTimeout(r, 100));
      }

      // 4) Fallback to defaults
      if (!cancelled) setSettings({ language: 'en', ramMB: 2048, maxRam: 16384 });
    };

    tryLoad();
    return () => { cancelled = true; };
  }, []);

  if (!settings) {
    return (
      <div style={{
        width: '100vw', height: '100vh',
        background: '#0F1115',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        color: '#4B5563', fontFamily: 'Inter, sans-serif',
        fontSize: '14px', fontWeight: 700, letterSpacing: '0.15em',
      }}>
        INITIALIZING...
      </div>
    );
  }

  return (
    <LanguageProvider initialLang={settings.language as 'ru' | 'en'}>
      <LauncherProvider>
        <DesktopLayout
          initialRamMB={settings.ramMB}
          initialMaxRam={settings.maxRam}
          onSettingsChange={(lang, ram) => {
            writeLS(lang, ram, settings.maxRam);
            if (window.go?.main?.App?.SaveSettings) {
              window.go.main.App.SaveSettings(lang, ram).catch(() => {});
            }
          }}
        >
          <PlayPage />
        </DesktopLayout>
      </LauncherProvider>
    </LanguageProvider>
  );
}
