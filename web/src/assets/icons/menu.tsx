import type { SVGProps } from "react";

type MenuIconProps = SVGProps<SVGSVGElement> & { open?: boolean };

export default function MenuIcon({ open = false, ...props }: MenuIconProps) {
    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            aria-hidden="true"
            {...props}
        >
            {open ? (
                <path d="M6 6l12 12M18 6L6 18" />
            ) : (
                <path d="M4 7h16M4 12h16M4 17h16" />
            )}
        </svg>
    );
}
