import DesktopLayout from './layouts/DesktopLayout';
import PlayPage from './pages/Play';
import { LauncherProvider } from './store/LauncherContext';
import { LanguageProvider } from './contexts/LanguageContext';
import './index.css';

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          SaveSettings?: (language: string, ramMB: number) => Promise<boolean>;
        };
      };
    };
  }
}

const LS_KEY = 'craftopia_settings';

function readLS(): { language: string; ramMB: number; maxRam: number } {
  try {
    const raw = localStorage.getItem(LS_KEY);
    if (raw) {
      const p = JSON.parse(raw);
      if (p && p.language) return p;
    }
  } catch (_) {}
  return { language: 'en', ramMB: 2048, maxRam: 16384 };
}

const initial = readLS();

export default function App() {
  return (
    <LanguageProvider initialLang={initial.language as 'ru' | 'en'}>
      <LauncherProvider>
        <DesktopLayout
          initialRamMB={initial.ramMB}
          initialMaxRam={initial.maxRam}
          onSettingsChange={(lang, ram) => {
            try {
              localStorage.setItem(LS_KEY, JSON.stringify({ language: lang, ramMB: ram, maxRam: initial.maxRam }));
            } catch (_) {}
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
