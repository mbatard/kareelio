import { Link } from 'react-router-dom';
import { LanguageToggle, ThemeToggle } from './Toggles';

export function PublicNavbar() {
  return (
    <nav className="w-full bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link to="/login" className="text-xl font-bold text-blue-600 dark:text-blue-400">
              Kareelio
            </Link>
          </div>
          <div className="flex items-center space-x-4">
            <LanguageToggle />
            <ThemeToggle />
          </div>
        </div>
      </div>
    </nav>
  );
}
