import { createContext, useContext, useState, ReactNode } from 'react';

export interface SetupWizardData {
  systemConfig?: {
    companyName: string;
    systemSettings: any;
  };
  adminData?: {
    name: string;
    email: string;
    password: string;
  };
}

interface SetupWizardContextType {
  wizardData: SetupWizardData;
  setSystemConfig: (config: SetupWizardData['systemConfig']) => void;
  setAdminData: (admin: SetupWizardData['adminData']) => void;
  clearWizardData: () => void;
}

const SetupWizardContext = createContext<SetupWizardContextType | undefined>(undefined);

interface SetupWizardProviderProps {
  children: ReactNode;
}

export function SetupWizardProvider({ children }: SetupWizardProviderProps) {
  const [wizardData, setWizardData] = useState<SetupWizardData>({});

  const setSystemConfig = (config: SetupWizardData['systemConfig']) => {
    setWizardData((prev) => ({
      ...prev,
      systemConfig: config,
    }));
  };

  const setAdminData = (admin: SetupWizardData['adminData']) => {
    setWizardData((prev) => ({
      ...prev,
      adminData: admin,
    }));
  };

  const clearWizardData = () => {
    setWizardData({});
  };

  const value: SetupWizardContextType = {
    wizardData,
    setSystemConfig,
    setAdminData,
    clearWizardData,
  };

  return <SetupWizardContext.Provider value={value}>{children}</SetupWizardContext.Provider>;
}

export function useSetupWizardContext(): SetupWizardContextType {
  const context = useContext(SetupWizardContext);
  if (context === undefined) {
    throw new Error('useSetupWizardContext must be used within a SetupWizardProvider');
  }
  return context;
}
