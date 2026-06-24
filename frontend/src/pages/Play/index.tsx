import { useState, useEffect } from 'react';
import { ChevronDown, ExternalLink } from 'lucide-react';
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

const MOCK_NEWS = [
  { id: 1, title: "Глобальное обновление 'Эра Света'", text: "Уважаемые игроки! Мы рады представить вам самое масштабное обновление этого года. Улучшена стабильность серверов, переработана система крафта и добавлено множество уникальных предметов. Мы долго работали над этим патчем и надеемся, что он вам понравится. В дополнение к этому мы подготовили специальный ивент, который начнется в эти выходные. Не упустите шанс получить уникальные награды за вход в игру и выполнение особых заданий! Кроме того, мы переработали баланс экономики, чтобы сделать процесс добычи золота более равномерным на всех этапах прокачки вашего персонажа.", date: "22 июня 2026", timeAgo: "5 минут назад" },
  { id: 2, title: "Запуск зимнего ивента", text: "Холодные ветра принесли с собой новые испытания! Собирайте уникальные ресурсы на заснеженных равнинах, получайте легендарные награды из ледяных сундуков и сражайтесь с ледяными драконами. Ивент продлится две недели. В рамках ивента будет доступна особая праздничная локация 'Ледяной Пик', где каждый день будут появляться новые цепочки квестов. За их выполнение вы получите очки зимней репутации, которые можно обменять на уникальные скины для вашего текущего комплекта брони и оружия.", date: "20 июня 2026", timeAgo: "2 дня назад" },
  { id: 3, title: "Свежий ребаланс классов", text: "В этом патче мы уделили особое внимание балансу. Маги стали значительно мощнее за счет новых заклинаний стихий, а лучникам добавили новые активные умения для повышения мобильности в бою на ближних дистанциях.", date: "15 июня 2026", timeAgo: "7 дней назад" },
  { id: 4, title: "Сезон арены открыт", text: "Скрестите клинки и приготовьте свои лучшие заклинания! Сражайтесь за звание лучшего гладиатора.", date: "10 июня 2026", timeAgo: "12 дней назад" },
  { id: 5, title: "Оптимизация сетевого кода", text: "Мы внимательно изучили ваши отзывы и внесли ряд критических исправлений в сетевой код для плавных сражений.", date: "5 июня 2026", timeAgo: "17 дней назад" },
  { id: 6, title: "Турнир по рыбной ловле!", text: "Любители мирных промыслов, настало ваше время! В эту субботу в главной столице пройдет грандиозный турнир по рыбной ловле с редкими наградами.", date: "1 июня 2026", timeAgo: "21 день назад" },
  { id: 7, title: "Новая система гильдий: Войны Альянсов", text: "Мы полностью переписали механику гильдий! Теперь вы можете объединять гильдии в Альянсы, объявлять войны за территории и строить собственные гильдейские замки. Строительство замков потребует работы всех участников гильдии, начиная с лесорубов и заканчивая архитекторами. Защита своего замка — это совершенно новый уровень PvP-боев, где огромную роль играют осадные орудия, тактика и слаженность всей гильдии. Первые битвы за земли начнутся уже на следующей неделе!", date: "28 мая 2026", timeAgo: "25 дней назад" },
  { id: 8, title: "Технические работы 25 мая", text: "Серверы будут недоступны с 02:00 до 04:00 по московскому времени для проведения планового технического обслуживания. Приносим извинения за неудобства.", date: "24 мая 2026", timeAgo: "29 дней назад" },
  { id: 9, title: "Итоги творческого конкурса 'Мир Craftopia'", text: "Конкурс артов и видеороликов успешно завершен! Мы получили более 500 потрясающих работ. Выбирать победителей было очень сложно. Спасибо всем за участие!", date: "20 мая 2026", timeAgo: "1 месяц назад" },
  { id: 10, title: "Новый мировой босс: Повелитель Бурь", text: "В пустошах проснулось древнее зло. Сразитесь с новым мировым боссом, обладающим разрушительными атаками стихии воздуха.", date: "15 мая 2026", timeAgo: "1 месяц назад" },
  { id: 11, title: "Введение системы ежедневных наград за вход", text: "Теперь каждый день, когда вы заходите в игру, вас ждет приятный бонус! Начиная от небольших сундуков с золотом и ресурсами в первый день, и заканчивая редкими ключами от подземелий при ежедневном входе в течение месяца.", date: "10 мая 2026", timeAgo: "1 месяц назад" },
  { id: 12, title: "Изменения в экономике: Алхимия", text: "Мы немного откорректировали стоимость зелий исцеления и маны. Это должно сбалансировать рынок.", date: "5 мая 2026", timeAgo: "1 месяц назад" },
  { id: 13, title: "Большое весеннее обновление уже здесь!", text: "Весна пришла в наш мир во всей красе! Деревья расцвели, а в магазинах появились новые весенние коллекции косметических предметов. Но это еще не всё — это одно из самых крупных обновлений этого полугодия! Мы добавили совершенно новую зону 'Цветущие Поля', где новичкам будет гораздо интереснее прокачиваться до 30 уровня. Появились новые питомцы и расширены слоты инвентаря для всех игроков.", date: "1 мая 2026", timeAgo: "1.5 месяца назад" },
  { id: 14, title: "Анонс: Грядущие изменения системы питомцев", text: "В скором времени система питомцев претерпит значительные изменения. Готовьтесь стать настоящими хозяевами зверей!", date: "25 апреля 2026", timeAgo: "2 месяца назад" },
  { id: 15, title: "Исправление ошибки с пропаданием квестовых предметов", text: "Важный хотфикс: исправлен редкий баг из-за которого квестовые предметы могли исчезать из инвентаря при переходе между локациями.", date: "20 апреля 2026", timeAgo: "2 месяцев назад" }
];

const dict = {
  ru: { readMore: "Читать далее", exit: "ВЫЙТИ", play: "ИГРАТЬ", wait: "ЖДИТЕ" },
  en: { readMore: "Read more", exit: "EXIT", play: "PLAY", wait: "WAIT" }
};

type PageState = 'idle' | 'busy' | 'error' | 'done' | 'updating';

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          StartRuntimePreparation?: () => void;
          KillMinecraft?: () => void;
          CheckForUpdate?: () => Promise<{available: boolean; version: string; sha256: string}>;
          ApplyUpdate?: () => Promise<boolean>;
        };
      };
    };
  }
}

export default function PlayPage() {
  const { lang, t, tPhase } = useLanguage();
  const b = dict[lang];
  const [isAtTop, setIsAtTop] = useState(true);
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

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    setIsAtTop(e.currentTarget.scrollTop < 10);
  };

  return (
    <div className="flex-1 flex flex-col h-full relative">
      <div className="flex-[2] px-8 pb-3 pt-8 min-h-0 flex flex-col relative overflow-hidden">
        <div onScroll={handleScroll}
          className="flex-1 overflow-y-auto space-y-4 pr-2 pb-16 [&::-webkit-scrollbar]:w-1 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:bg-white/10 hover:[&::-webkit-scrollbar-thumb]:bg-[#94A3B8]/20 [&::-webkit-scrollbar-thumb]:rounded-full"
        >
          {MOCK_NEWS.map((news) => (
            <div key={news.id}
              className="bg-[#1C1F26] border border-white/[0.03] hover:border-white/[0.08] rounded-lg p-5 flex flex-col cursor-pointer"
            >
              <div className="flex items-center gap-3 mb-1.5 flex-wrap">
                <h1 className="text-white text-[15px] font-bold tracking-tight leading-tight">{news.title}</h1>
                <span className="text-[12px] text-[#4B5563] font-medium">
                  {news.date} <span className="opacity-70 text-[11px] ml-1">({news.timeAgo})</span>
                </span>
              </div>
              <div className="relative">
                <p className={`text-[#4B5563] text-[13px] leading-relaxed ${news.text.length > 250 ? 'line-clamp-3' : ''}`}>{news.text}</p>
                {news.text.length > 250 && (
                  <div className="absolute bottom-0 right-0 z-10 flex items-center justify-end pl-10 bg-[#1C1F26]">
                    <div className="absolute left-[-32px] bottom-0 w-[32px] h-full bg-[#1C1F26]" />
                    <span className="text-[#94A3B8] text-[13px] font-medium transition-all duration-200 cursor-pointer select-none leading-relaxed flex items-center justify-end whitespace-nowrap hover:text-white hover:scale-[1.03] active:scale-[0.97] origin-right inline-flex">
                      {b.readMore}
                      <ExternalLink className="w-3.5 h-3.5 ml-1 inline text-current" strokeWidth={2} />
                    </span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>

        {isAtTop && (
          <div className="absolute bottom-16 left-1/2 -translate-x-1/2 flex flex-col items-center pointer-events-none z-25 transition-opacity duration-300">
            <div className="flex flex-col items-center justify-center w-10 h-10 bg-[#0F1115]/90 rounded-full border border-white/[0.06] backdrop-blur-sm shadow-lg animate-float">
              <ChevronDown className="w-6 h-6 text-[#94A3B8]" strokeWidth={2.5} />
            </div>
          </div>
        )}

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
