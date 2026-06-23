import { useState, useRef, useEffect } from 'react';
import { Settings, Power, LifeBuoy, Check, RotateCcw, X } from 'lucide-react';
import { useLanguage } from '../contexts/LanguageContext';
import { useLauncher } from '../store/LauncherContext';

declare global {
  interface Window {
    runtime?: { Quit?: () => void; };
    go?: {
      main?: {
        App?: {
          GetSettings?: () => Promise<[string, number, number]>;
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

interface DesktopLayoutProps {
  children: React.ReactNode;
  initialRamMB?: number;
  initialMaxRam?: number;
  onSettingsChange?: (language: string, ramMB: number) => void;
}

export default function DesktopLayout({ children, initialRamMB, initialMaxRam, onSettingsChange }: DesktopLayoutProps) {
  const { lang, setLang, t } = useLanguage();
  const { resetLauncher } = useLauncher();

  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [ram, setRam] = useState(initialRamMB || 2048);
  const [maxRam, setMaxRam] = useState(initialMaxRam || 16384);
  const [isConfirmResetOpen, setIsConfirmResetOpen] = useState(false);
  const [isResetAnimating, setIsResetAnimating] = useState(false);
  const [showSuccessToast, setShowSuccessToast] = useState(false);

  const sliderRef = useRef<HTMLInputElement>(null);

  const saveSettings = (newLang: string, newRam: number) => {
    if (onSettingsChange) onSettingsChange(newLang, newRam);
  };

  const handleClose = () => {
    // Kill Minecraft if running
    if (window.go?.main?.App?.KillMinecraft) {
      try { window.go.main.App.KillMinecraft(); } catch (_) {}
    }
    // Quit app
    if (window.runtime?.Quit) {
      window.runtime.Quit();
    } else {
      console.log('[Wails Mock]: Quit');
    }
  };

  const handleWheel = (e: React.WheelEvent) => {
    if (document.activeElement === sliderRef.current || e.currentTarget === sliderRef.current?.parentElement) {
      if (e.deltaY < 0) {
        setRam(prev => {
          const next = Math.min(maxRam, prev + 256);
          saveSettings(lang, next);
          return next;
        });
      } else {
        setRam(prev => {
          const next = Math.max(2048, prev - 256);
          saveSettings(lang, next);
          return next;
        });
      }
    }
  };

  const handleResetLaunchEngine = () => {
    setIsResetAnimating(true);
    if (window.go?.main?.App?.ResetLauncher) {
      window.go.main.App.ResetLauncher().then((ok: boolean) => {
        setIsResetAnimating(false);
        setIsConfirmResetOpen(false);
        if (ok) {
          setShowSuccessToast(true);
          setTimeout(() => setShowSuccessToast(false), 3000);
        }
      }).catch(() => {
        setIsResetAnimating(false);
        setIsConfirmResetOpen(false);
      });
    } else {
      setTimeout(() => {
        resetLauncher();
        setIsResetAnimating(false);
        setIsConfirmResetOpen(false);
        setShowSuccessToast(true);
        setTimeout(() => setShowSuccessToast(false), 3000);
      }, 1500);
    }
  };

  return (
    <div className="h-screen w-screen bg-bg-deep flex items-center justify-center overflow-hidden relative">
      {/* Background pattern — only in browser preview */}
      {!window.runtime?.Quit && (
        <div className="absolute inset-0 bg-[radial-gradient(#2A2D34_1px,transparent_1px)] [background-size:32px_32px] opacity-10 pointer-events-none" />
      )}

      {/* Контейнер-Окно — заполняет окно Wails полностью */}
      <div
        id="wails-window-frame"
        className="w-full h-full bg-bg-deep border border-white/5 rounded-xl shadow-[0_40px_100px_rgba(0,0,0,0.85)] flex flex-col overflow-hidden relative z-10"
      >
        {/* Header Top Navigation */}
        <header
          className="h-[54px] bg-[#0B0D10] border-b border-[#1C1F26] flex items-center justify-between px-6 z-20 shrink-0 relative"
        >
          <div className="flex items-center gap-2">
            <span className="text-white font-black text-[12px] tracking-[0.25em] uppercase italic bg-gradient-to-r from-white via-white/90 to-white/50 bg-clip-text text-transparent select-none font-sans">
              CraftopiaMC Launcher
            </span>
          </div>

          <div className="flex items-center gap-2">
            <a
              href="https://t.me/craftopiamc"
              target="_blank"
              rel="noopener noreferrer"
              className="w-[34px] h-[34px] flex items-center justify-center text-[#4B5563] hover:text-white hover:bg-white/5 rounded-lg transition-all cursor-pointer relative group"
            >
              <LifeBuoy className="w-[18px] h-[18px] transition-all group-hover:rotate-12" strokeWidth={2} />
              <div className="absolute right-0 top-[calc(100%+8px)] opacity-0 group-hover:opacity-100 transition-opacity duration-150 ease-in-out bg-[#1C1F26] text-white text-[10px] py-1 px-2.5 rounded-md shadow-[0_4px_24px_rgba(0,0,0,0.5)] border border-white/[0.04] pointer-events-none z-50 whitespace-nowrap font-medium flex items-center uppercase tracking-wider">
                {t.sidebarSupport}
              </div>
            </a>

            <button
              onClick={() => setIsSettingsOpen(true)}
              className="w-[34px] h-[34px] flex items-center justify-center text-[#4B5563] hover:text-[#94A3B8] hover:bg-white/5 rounded-lg transition-all cursor-pointer relative group"
            >
              <Settings className="w-[18px] h-[18px] transition-all group-hover:rotate-45" strokeWidth={2} />
              <div className="absolute right-0 top-[calc(100%+8px)] opacity-0 group-hover:opacity-100 transition-opacity duration-150 ease-in-out bg-[#1C1F26] text-white text-[10px] py-1 px-2.5 rounded-md shadow-[0_4px_24px_rgba(0,0,0,0.5)] border border-white/[0.04] pointer-events-none z-50 whitespace-nowrap font-medium flex items-center uppercase tracking-wider">
                {t.sidebarSettings}
              </div>
            </button>

            <button
              onClick={handleClose}
              className="w-[34px] h-[34px] flex items-center justify-center text-[#4B5563] hover:text-red-500 hover:bg-red-500/10 rounded-lg transition-all cursor-pointer relative group"
            >
              <Power className="w-[18px] h-[18px] transition-transform group-hover:scale-105" strokeWidth={2.5} />
              <div className="absolute right-0 top-[calc(100%+8px)] opacity-0 group-hover:opacity-100 transition-opacity duration-150 ease-in-out bg-[#1C1F26] text-white text-[10px] py-1 px-2.5 rounded-md shadow-[0_4px_24px_rgba(0,0,0,0.5)] border border-white/[0.04] pointer-events-none z-50 whitespace-nowrap font-medium flex items-center uppercase tracking-wider">
                {t.sidebarExit}
              </div>
            </button>
          </div>
        </header>

        {/* Рабочая область приложения */}
        <main className="flex-1 flex flex-col relative overflow-hidden bg-bg-deep">
          <div
            className="flex-1 flex flex-col min-h-0 relative z-10"
          >
            {children}
          </div>
        </main>

        {/* Settings Modal Overlay */}
        <div
          className={`absolute inset-0 bg-black/65 flex items-center justify-center z-50 p-6 transition-all duration-200 ${isSettingsOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'}`}
          style={{ willChange: 'transform, opacity' } as React.CSSProperties}
        >
          <div
            className={`w-full max-w-[420px] bg-[#0B0D10] border border-white/[0.08] rounded-2xl p-7 shadow-2xl relative flex flex-col gap-6 overflow-hidden transition-all duration-200 transform ${isSettingsOpen ? 'scale-100 translate-y-0 opacity-100' : 'scale-95 translate-y-4 opacity-0'}`}
          >
            <div className="flex items-center justify-between border-b border-white/[0.06] pb-4">
              <span className="text-[15px] font-black tracking-[0.15em] text-white uppercase italic">
                {t.settingsTitle}
              </span>
              <button
                onClick={() => setIsSettingsOpen(false)}
                className="w-8 h-8 rounded-md flex items-center justify-center text-[#4B5563] hover:text-white hover:bg-white/5 transition-colors cursor-pointer"
              >
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="flex flex-col gap-5">
              <div className="flex flex-col gap-2">
                <label className="text-[13px] font-black uppercase tracking-wider text-[#94A3B8]">
                  {t.languageLabel}
                </label>
                <div className="flex bg-[#1C1F26] p-1 rounded-lg border border-white/[0.04]">
                  <button
                    onClick={() => { setLang('ru'); saveSettings('ru', ram); }}
                    className={`flex-1 py-2.5 rounded-md text-[13px] font-black uppercase transition-all cursor-pointer ${lang === 'ru' ? 'bg-[#0F1115] text-white shadow-sm ring-1 ring-white/[0.02]' : 'text-[#64748B] hover:text-[#94A3B8]'}`}
                  >
                    Русский
                  </button>
                  <button
                    onClick={() => { setLang('en'); saveSettings('en', ram); }}
                    className={`flex-1 py-2.5 rounded-md text-[13px] font-black uppercase transition-all cursor-pointer ${lang === 'en' ? 'bg-[#0F1115] text-white shadow-sm ring-1 ring-white/[0.02]' : 'text-[#64748B] hover:text-[#94A3B8]'}`}
                  >
                    English
                  </button>
                </div>
              </div>

              <div className="flex flex-col gap-2">
                <div className="flex justify-between items-center text-[13px] font-black uppercase tracking-wider">
                  <span className="text-[#94A3B8]">{t.ramAllocation}</span>
                  <span className="text-white font-mono font-black text-sm tracking-wide">
                    {(ram / 1024).toFixed(1)} GB <span className="text-[#64748B]">({ram} MB)</span>
                  </span>
                </div>

                <div
                  className="group relative flex flex-col gap-2 pt-2"
                  onWheel={handleWheel}
                >
                  <input
                    ref={sliderRef}
                    type="range"
                    min={2048}
                    max={maxRam}
                    step={256}
                    value={ram}
                    onChange={(e) => {
                      const v = Number(e.target.value);
                      setRam(v);
                      saveSettings(lang, v);
                    }}
                    className="w-full h-2 bg-[#1C1F26] rounded-full appearance-none cursor-pointer focus:outline-none focus:ring-1 focus:ring-white/10 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-[18px] [&::-webkit-slider-thumb]:h-[18px] [&::-webkit-slider-thumb]:bg-white [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:hover:scale-110 [&::-webkit-slider-thumb]:transition-transform"
                    style={{
                      background: maxRam > 2048
                        ? `linear-gradient(to right, rgba(255,255,255,0.85) ${((ram - 2048) / (maxRam - 2048)) * 100}%, #1C1F26 ${((ram - 2048) / (maxRam - 2048)) * 100}%)`
                        : 'bg-[#1C1F26]'
                    }}
                  />
                </div>
              </div>

              <div className="pt-4 border-t border-white/[0.06] mt-1">
                <button
                  onClick={() => setIsConfirmResetOpen(true)}
                  className="w-full h-12 bg-red-500/10 hover:bg-red-500/15 border border-red-500/20 text-red-500 font-black text-[13px] tracking-widest uppercase rounded-lg flex items-center justify-center gap-2 transition-all cursor-pointer active:scale-98"
                >
                  <RotateCcw className="w-4 h-4" />
                  {t.reinstallLauncher}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Reset Confirmation Overlay */}
        <div
          className={`absolute inset-0 bg-black/80 flex items-center justify-center z-[60] p-6 transition-all duration-200 ${isConfirmResetOpen ? 'opacity-100 pointer-events-auto' : 'opacity-0 pointer-events-none'}`}
          style={{ willChange: 'transform, opacity' } as React.CSSProperties}
        >
          <div
            className={`w-full max-w-sm bg-[#0F1115] border border-white/[0.06] rounded-xl p-7 shadow-2xl flex flex-col gap-5 text-center items-center transition-all duration-200 transform ${isConfirmResetOpen ? 'scale-100 translate-y-0 opacity-100' : 'scale-95 translate-y-4 opacity-0'}`}
          >
            <div className="w-12 h-12 bg-red-500/10 rounded-full flex items-center justify-center text-red-500 mb-1 border border-red-500/20">
              <RotateCcw className="w-6 h-6 animate-spin" style={{ animationDuration: '3s' }} />
            </div>
            <h3 className="text-white text-base font-black tracking-wider uppercase italic">
              {t.reinstallLauncher}
            </h3>
            <p className="text-[#94A3B8] text-[14px] font-bold leading-relaxed">
              {t.reinstallConfirm}
            </p>

            <div className="flex gap-3 w-full mt-2">
              <button
                onClick={() => setIsConfirmResetOpen(false)}
                className="flex-1 h-11 bg-white/5 border border-white/5 hover:bg-white/10 text-white font-black text-[13px] tracking-widest uppercase rounded-lg transition-all cursor-pointer animate-none"
                disabled={isResetAnimating}
              >
                {lang === 'ru' ? 'ОТМЕНА' : 'CANCEL'}
              </button>
              <button
                onClick={handleResetLaunchEngine}
                className="flex-1 h-11 bg-red-500 hover:bg-red-600 text-white font-black text-[13px] tracking-widest uppercase rounded-lg transition-all cursor-pointer flex items-center justify-center"
                disabled={isResetAnimating}
              >
                {isResetAnimating ? (
                  <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                ) : (
                  lang === 'ru' ? 'СБРОСИТЬ' : 'RESET'
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Floating Success Toast */}
        <div
          className={`absolute top-16 left-1/2 -translate-x-1/2 z-[70] bg-[#1C1F26] border border-green-500/20 px-4 py-2.5 rounded-lg flex items-center gap-2 shadow-xl transition-all duration-200 transform ${showSuccessToast ? 'opacity-100 translate-y-0 scale-100' : 'opacity-0 -translate-y-4 scale-95 pointer-events-none'}`}
          style={{ willChange: 'transform, opacity' } as React.CSSProperties}
        >
          <Check className="w-4 h-4 text-green-500 shrink-0" strokeWidth={2.5} />
          <span className="text-white text-xs font-bold uppercase tracking-wider">{t.reinstalledSuccess}</span>
        </div>
      </div>
    </div>
  );
}
