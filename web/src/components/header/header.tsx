import css from "./header.module.scss";
import cart from "./cart.module.scss";
import profile from "./profile.module.scss";
import search from "./search.module.scss";
import CartIcon from "@/assets/icons/cart";
import StarIcon from "@/assets/icons/star";
import { useEffect, useRef } from "react";

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
            <div className={css.content}>
                <nav className={css.navbar_container}>
                    <ul className={css.navbar_list}>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                <h1 className={css.title}>Waystore</h1>
                            </a>
                        </li>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                Popular
                            </a>
                        </li>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                Library
                            </a>
                        </li>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                Merchant
                            </a>
                        </li>
                    </ul>
                </nav>
                <div className={search.container}>
                    <input
                        ref={searchRef}
                        type="text"
                        className={search.input_bar}
                        placeholder="What are you looking for?"
                    />
                    <div className={search.slash}>/</div>
                </div>
                <nav className={`${css.navbar_container} ${css.justify_right}`}>
                    <ul className={css.navbar_list}>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                <StarIcon className={css.link_icon} />
                                Wishlist
                            </a>
                        </li>
                        <li className={css.navbar_item}>
                            <a href="" className={css.navbar_link}>
                                <div className={cart.container}>
                                    <CartIcon className={cart.icon} />
                                    Cart 0
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
                                            Settings
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
        </header>
    );
}
