import { Dropdown } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { useAuthStore, SUPPORTED_LOCALES } from '@gastrack/shared';
import type { MenuProps } from 'antd';
import type { CSSProperties } from 'react';

interface LanguageSwitcherProps {
  /** 图标样式覆盖 */
  style?: CSSProperties;
}

export default function LanguageSwitcher({ style }: LanguageSwitcherProps) {
  const { i18n } = useTranslation();
  const { user, updateProfile } = useAuthStore();

  const languageItems: MenuProps['items'] = SUPPORTED_LOCALES.map((l) => ({
    key: l.value,
    label: l.label,
  }));

  const handleLanguageChange: MenuProps['onClick'] = async ({ key }) => {
    await i18n.changeLanguage(key);
    localStorage.setItem('locale', key);
    // 已登录时同步保存到后端
    if (user) {
      try {
        await updateProfile({ locale: key });
      } catch {
        // 静默失败，前端语言已切换
      }
    }
  };

  return (
    <Dropdown
      menu={{ items: languageItems, onClick: handleLanguageChange }}
      placement="bottomRight"
    >
      <GlobalOutlined
        style={{
          fontSize: 18,
          cursor: 'pointer',
          verticalAlign: 'middle',
          ...style,
        }}
      />
    </Dropdown>
  );
}
