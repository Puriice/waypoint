import type { SVGProps } from "react";

export default function CartIcon(props: SVGProps<SVGSVGElement>) {
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
            strokeLinejoin="round"
            aria-hidden="true"
            {...props}
        >
            {/* Handle */}
            <path d="M1 1h4l1 5" />
            {/* Basket, filled */}
            <path
                d="M6 6h17l-1.6 8.4a2 2 0 0 1-2 1.6H9.68a2 2 0 0 1-2-1.61Z"
                fill="currentColor"
            />
            {/* Wheels, filled */}
            <circle cx="9" cy="21" r="2" fill="currentColor" stroke="none" />
            <circle cx="20" cy="21" r="2" fill="currentColor" stroke="none" />
        </svg>
    );
}
