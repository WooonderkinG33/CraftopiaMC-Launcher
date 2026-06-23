import { createContext, useContext } from 'react';

interface LauncherContextType {
  resetLauncher: () => void;
}

const LauncherContext = createContext<LauncherContextType>({
  resetLauncher: () => {},
});

export function LauncherProvider({ children }: { children: React.ReactNode }) {
  const resetLauncher = () => {
    console.log('[Launcher] Reset requested');
  };

  return (
    <LauncherContext.Provider value={{ resetLauncher }}>
      {children}
    </LauncherContext.Provider>
  );
}

export function useLauncher() {
  return useContext(LauncherContext);
}
