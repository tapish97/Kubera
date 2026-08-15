type IconName = "home" | "inventory" | "plus" | "sale" | "more" | "arrow" | "sparkle" | "trend" | "box";

export function DashboardIcon({ name, className = "h-5 w-5" }: { name: IconName; className?: string }) {
  const paths: Record<IconName, React.ReactNode> = {
    home: <path d="m3 10 9-7 9 7v9a2 2 0 0 1-2 2h-4v-7H9v7H5a2 2 0 0 1-2-2Z" />,
    inventory: <><path d="M4 7.5 12 3l8 4.5v9L12 21l-8-4.5Z" /><path d="m4 7.5 8 4.5 8-4.5M12 12v9" /></>,
    plus: <><path d="M12 5v14M5 12h14" /></>,
    sale: <><path d="M4 5h16v14H4z" /><path d="M8 9h8M8 13h5" /></>,
    more: <><circle cx="5" cy="12" r="1" /><circle cx="12" cy="12" r="1" /><circle cx="19" cy="12" r="1" /></>,
    arrow: <><path d="M5 12h14M14 7l5 5-5 5" /></>,
    sparkle: <><path d="m12 3 1.4 4.1L17.5 8.5l-4.1 1.4L12 14l-1.4-4.1-4.1-1.4 4.1-1.4Z" /><path d="m18.5 15 .7 2.3 2.3.7-2.3.7-.7 2.3-.7-2.3-2.3-.7 2.3-.7Z" /></>,
    trend: <><path d="m4 16 5-5 4 4 7-8" /><path d="M15 7h5v5" /></>,
    box: <><path d="M4 7h16v13H4zM7 4h10v3" /><path d="M9 11h6" /></>,
  };

  return <svg aria-hidden="true" className={className} fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">{paths[name]}</svg>;
}
