import css from "./header.module.scss";
import cart from "./cart.module.scss";
import profile from "./profile.module.scss";
import search from "./search.module.scss";
import CartIcon from "@/assets/icons/cart";
import BooksIcon from "@/assets/icons/books";
import MenuIcon from "@/assets/icons/menu";
import SearchIcon from "@/assets/icons/search";
import { useEffect, useRef } from "react";
import { Form } from "react-router";

const PRIMARY_LINKS = [
    { label: "Popular", href: "" },
    { label: "Wishlist", href: "" },
];

export default function Header() {
    const searchRef = useRef<HTMLInputElement>(null);

    useEffect(() => {
        function onKeyDown(e: KeyboardEvent) {
            if (e.key !== "/" || e.ctrlKey || e.metaKey || e.altKey) return;
            const target = e.target as HTMLElement | null;

            const isTextInput =
                target instanceof HTMLInputElement &&
                !["checkbox", "radio", "button", "submit", "reset"].includes(
                    target.type,
                );
            if (
                target &&
                (target.isContentEditable ||
                    isTextInput ||
                    ["TEXTAREA", "SELECT"].includes(target.tagName))
            ) {
                return;
            }
            e.preventDefault();
            searchRef.current?.focus();
        }
        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    }, []);

    return (
        <header className={css.root}>
            <input
                type="checkbox"
                id="menu_toggle"
                className={css.menu_toggle}
                aria-label="Menu"
                aria-controls="mobile_menu"
            />
            <div className={css.content}>
                <label htmlFor="menu_toggle" className={css.menu_btn}>
                    <MenuIcon className={css.menu_icon_closed} />
                    <MenuIcon open className={css.menu_icon_open} />
                </label>
                <nav className={`${css.navbar_container} ${css.primary_nav}`}>
                    <ul className={css.navbar_list}>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                <h1 className={css.title}>Waystore</h1>
                            </a>
                        </li>
                        {PRIMARY_LINKS.map((link) => (
                            <li
                                key={link.label}
                                className={`${css.navbar_item} ${css.desktop_only}`}
                            >
                                <a href={link.href} className={css.navbar_link}>
                                    {link.label}
                                </a>
                            </li>
                        ))}
                    </ul>
                </nav>
                <Form
                    method="get"
                    action="/search"
                    role="search"
                    className={search.container}
                >
                    <div className={search.box}>
                        <input
                            ref={searchRef}
                            type="search"
                            name="q"
                            aria-label="Search"
                            className={search.input_bar}
                            placeholder="What are you looking for?"
                        />
                        <div className={search.slash}>/</div>
                    </div>
                    <button
                        type="submit"
                        className={search.button}
                        aria-label="Search"
                    >
                        <SearchIcon className={search.button_icon} />
                    </button>
                </Form>
                <nav className={`${css.navbar_container} ${css.justify_right}`}>
                    <ul className={css.navbar_list}>
                        <li className={css.navbar_item}>
                            <a
                                href=""
                                className={css.navbar_link}
                                aria-label="Library"
                            >
                                <BooksIcon className={css.link_icon} />
                                <span className={css.label}>Library</span>
                            </a>
                        </li>
                        <li className={css.navbar_item}>
                            <a
                                href=""
                                className={css.navbar_link}
                                aria-label="Cart, 0 items"
                            >
                                <div className={cart.container}>
                                    <CartIcon className={cart.icon} />
                                    <span className={css.label}>Cart </span>0
                                </div>
                            </a>
                        </li>
                        <li className={css.navbar_item}>
                            <div className={profile.container}>
                                <label
                                    htmlFor="profile_menu"
                                    className={profile.btn}
                                >
                                    <img
                                        src=""
                                        alt=""
                                        className={profile.picture}
                                    />
                                </label>
                                <input
                                    type="checkbox"
                                    id="profile_menu"
                                    className={profile.checkbox}
                                    aria-label="Profile Menu"
                                    aria-controls="profile_menu"
                                />
                                <ul className={profile.menus}>
                                    <li className={profile.menu_item}>
                                        <div className={profile.detail}>
                                            <img
                                                src=""
                                                alt=""
                                                className={profile.picture}
                                            />
                                            <div className={profile.name}>
                                                Purinutt Amartayavis
                                            </div>
                                            <a href="" className={profile.link}>
                                                Go to profile
                                            </a>
                                        </div>
                                    </li>
                                    <li className={profile.menu_item}>
                                        <a
                                            href=""
                                            className={profile.profile_link}
                                        >
                                            Merchants
                                        </a>
                                    </li>
                                    <li className={profile.menu_item}>
                                        <a
                                            href=""
                                            className={profile.profile_link}
                                        >
                                            Settings
                                        </a>
                                    </li>
                                    <li className={profile.menu_item}>
                                        <a
                                            href=""
                                            className={profile.profile_link}
                                        >
                                            Help
                                        </a>
                                    </li>
                                    <li className={profile.menu_item}>
                                        <a
                                            href=""
                                            className={profile.profile_link}
                                        >
                                            Sign out
                                        </a>
                                    </li>
                                </ul>
                            </div>
                        </li>
                    </ul>
                </nav>
            </div>
            <nav id="mobile_menu" className={css.mobile_menu}>
                <ul className={css.mobile_list}>
                    {PRIMARY_LINKS.map((link) => (
                        <li key={link.label}>
                            <a href={link.href} className={css.mobile_link}>
                                {link.label}
                            </a>
                        </li>
                    ))}
                </ul>
            </nav>
        </header>
    );
}
