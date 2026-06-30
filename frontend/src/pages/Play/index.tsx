import { useState, useEffect } from 'react';
import { ChevronDown } from 'lucide-react';
import { useLanguage } from '../../contexts/LanguageContext';

declare global {
  interface Window {
    runtime?: {
      EventsOn?: (event: string, cb: (...args: any[]) => void) => void;
    };
    go?: {
      main?: {
        App?: {
          StartRuntimePreparation?: () => void;
          KillMinecraft?: () => void;
        };
      };
    };
  }
}

const dict = {
  ru: { exit: "ВЫЙТИ", play: "ИГРАТЬ", wait: "ЖДИТЕ" },
  en: { exit: "EXIT", play: "PLAY", wait: "WAIT" }
};

type PageState = 'idle' | 'busy' | 'error' | 'done' | 'updating';

export default function PlayPage() {
  const { lang, t, tPhase } = useLanguage();
  const b = dict[lang];
  const [pageState, setPageState] = useState<PageState>('idle');
  const [progress, setProgress] = useState(0);
  const [statusText, setStatusText] = useState('');
  const [speed, setSpeed] = useState('');
  const [errorCountdown, setErrorCountdown] = useState(0);
  const [hasError, setHasError] = useState(false);

  useEffect(() => {
    if (window.runtime?.EventsOn) {
      window.runtime.EventsOn('runtimeStatus', (msg: string, pct: number, spd: string, _dl: number, _total: number) => {
        if (pct === -1) {
          setPageState('error');
          setHasError(true);
          const countdown = spd === 'ERROR' ? _dl : 0;
          setStatusText(tPhase(msg) + ' ' + countdown);
          setErrorCountdown(countdown);
        } else if (pct === -2) {
          setPageState('error');
          setHasError(true);
          setStatusText(tPhase(msg));
          setProgress(0);
          setSpeed('');
          setTimeout(() => { setPageState('idle'); setProgress(0); }, 3000);
        } else if (pct === 100 && msg === 'gameRunning') {
          setPageState('done');
          setProgress(100);
          setStatusText(tPhase(msg));
          setSpeed('');
          setHasError(false);
          setErrorCountdown(0);
        } else if (msg === 'updating') {
          setPageState('updating');
          setProgress(pct);
          setStatusText(tPhase('updating', pct));
          setSpeed('');
          setHasError(false);
          setErrorCountdown(0);
        } else if (msg === 'updateDone') {
          setPageState('updating');
          setProgress(100);
          setStatusText('✓ ' + tPhase('updateDone'));
          setSpeed('');
          setHasError(false);
          setErrorCountdown(0);
        } else if ((pct === 0 && msg === 'MC_EXITED') || (pct === 0 && msg === 'MC_KILLED')) {
          setPageState('idle');
          setProgress(0);
          setStatusText('');
          setSpeed('');
          setHasError(false);
          setErrorCountdown(0);
        } else {
          setPageState('busy');
          setProgress(pct);
          setStatusText(tPhase(msg, pct));
          setSpeed(spd || '');
          setHasError(false);
          setErrorCountdown(0);
        }
      });
    }
  }, [tPhase]);

  const handleLaunchToggle = () => {
    if (pageState === 'done') {
      if (window.go?.main?.App?.KillMinecraft) {
        window.go.main.App.KillMinecraft();
      }
      return;
    }
    if (hasError || pageState === 'idle') {
      setPageState('busy');
      setProgress(0);
      setStatusText(t.statusReady);
      if (window.go?.main?.App?.StartRuntimePreparation) {
        window.go.main.App.StartRuntimePreparation();
      } else {
        setTimeout(() => {
          setPageState('done');
          setProgress(100);
          setStatusText('MOCK: GAME RUNNING');
        }, 2000);
      }
    }
  };

  const getProgressColor = () => {
    if (pageState === 'updating') return "bg-[#22C55E]";
    if (hasError) return "bg-[#EF4444]";
    if (pageState === 'done') return "bg-[#94A3B8]";
    if (pageState === 'idle') return "bg-[#4B5563]";
    return "bg-[#94A3B8]";
  };

  return (
    <div className="flex-1 flex flex-col h-full relative">
      {/* News placeholder */}
      <div className="flex-[2] px-8 pb-3 pt-8 min-h-0 flex flex-col relative overflow-hidden">
        <div className="flex-1 flex items-center justify-center">
          <span className="text-[#4B5563] text-[14px] font-bold tracking-wider uppercase opacity-60 select-none">
            {t.newsPlaceholder}
          </span>
        </div>
        <div className="absolute bottom-0 left-8 right-8 h-12 bg-[#0F1115] pointer-events-none z-20" />
      </div>

      <div className="shrink-0 relative z-50 px-8 pb-8 pt-0">
        <div className="bg-[#1C1F26] border border-white/[0.04] rounded-lg p-4.5 relative overflow-visible flex flex-row items-center gap-6 shadow-xl">

          <div className="flex-1 flex flex-col justify-center relative z-10 pl-1">
            <div className="flex items-end justify-between mb-1.5">
              <span
                className={`font-black text-[12px] tracking-[0.15em] uppercase truncate transition-colors duration-1000 ${hasError ? 'text-red-500' : pageState === 'done' ? 'text-[#94A3B8]' : pageState === 'idle' ? 'text-[#4B5563]' : 'text-white'}`}
              >
                {statusText || t.statusReady}
              </span>
            </div>

            <div className="w-full h-2.5 bg-black/40 rounded-full overflow-hidden flex shadow-inner">
              <div
                className={`h-full ${getProgressColor()} rounded-full transition-all duration-700 ease-out`}
                style={{ width: pageState === 'idle' ? '0%' : `${progress}%` }}
              />
            </div>

            {pageState === 'busy' && speed && (
              <div className="mt-1.5 text-[11px] font-semibold text-[#4B5563] tracking-wide">{speed}</div>
            )}
          </div>

          <div className="shrink-0 relative z-10 group flex items-center justify-center">
            <button
              onClick={handleLaunchToggle}
              className={`w-[132px] h-[40px] rounded-md text-center select-none tracking-[0.25em] font-extrabold text-[13px] transition-all duration-300 cursor-pointer uppercase shadow-md flex items-center justify-center ${
                hasError
                  ? 'bg-white text-[#0F1115] hover:bg-[#94A3B8] hover:text-white active:scale-[0.98]'
                  : pageState === 'done'
                    ? 'bg-red-500/10 border border-red-500/30 hover:bg-red-500/20 text-red-100 active:scale-[0.98]'
                    : pageState !== 'idle'
                      ? 'bg-white/5 border border-white/5 text-[#4B5563] cursor-pointer text-[12px] hover:bg-white/10'
                      : 'bg-white text-[#0F1115] hover:bg-[#94A3B8] hover:text-white active:scale-[0.98]'
              }`}
            >
              {pageState === 'done' ? b.exit : pageState === 'idle' ? b.play : b.wait}
            </button>
          </div>

        </div>
      </div>

    </div>
  );
}
