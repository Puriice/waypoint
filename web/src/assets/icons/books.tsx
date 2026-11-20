import { useId, type SVGProps } from "react";

export default function BooksIcon(props: SVGProps<SVGSVGElement>) {
    // Unique per instance so several icons on a page don't share a mask
    const maskId = useId();

    return (
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
            {...props}
        >
            {/* Spine bands are cut out of the books, so the page shows through */}
            <mask id={maskId}>
                <rect width="24" height="24" fill="white" />
                <g stroke="black" strokeWidth="1.5">
                    <path d="M1 8h7" />
                    <path d="M6 11h7" />
                    <path d="M16 11h8" transform="rotate(-15 17.3 22)" />
                </g>
            </mask>
            <g mask={`url(#${maskId})`}>
                {/* Two upright books */}
                <rect x="1" y="2" width="6" height="20" rx="1" />
                <rect
                    x="7.5"
                    y="5"
                    width="5.5"
                    height="17"
                    rx="1"
                    opacity="0.75"
                />
                {/* Third book resting on its bottom-left corner, leaning on the second */}
                <rect
                    x="17.3"
                    y="5.5"
                    width="5.5"
                    height="16.5"
                    rx="1"
                    opacity="0.55"
                    transform="rotate(-15 17.3 22)"
                />
            </g>
        </svg>
    );
}
