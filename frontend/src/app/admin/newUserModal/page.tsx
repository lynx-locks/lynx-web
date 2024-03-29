"use client";

import Modal from "@/components/modal/modal";
import { useContext, useEffect, useState } from "react";
import styles from "../modals.module.css";
import SearchDropdown from "@/components/searchDropdown/searchDropdown";
import { Options, SelectType } from "@/types/selectOptions";
import { SubmitButton } from "@/components/button/button";
import { getRoleOptions } from "@/data/roles";
import { useRouter } from "next/navigation";
import axios from "@/axios/client";
import { AdminContext } from "../layout";

const emailRegex = /^\w+([.-]?\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/;

export default function NewUserModal() {
  const router = useRouter();
  const [newUser, setNewUser] = useState<{ name: string; email: string }>({
    name: "",
    email: "",
  });
  const [selectedRoleOption, setSelectedRoleOption] =
    useState<SelectType>(null);
  const [roles, setRoles] = useState<Options[]>([]);
  const [isAdmin, setIsAdmin] = useState(false);
  const [disabled, setDisabled] = useState(false);
  const { users, setUsers } = useContext(AdminContext);

  useEffect(() => {
    const f = async () => {
      setRoles(await getRoleOptions());
    };
    f();
  }, []);

  const handleModalClose = () => {
    router.push("/admin");
  };

  const handleModalSubmit = async () => {
    setDisabled(true);
    // handle adding new user
    if (emailRegex.test(newUser.email)) {
      const userResp = await axios.post("/users", {
        name: newUser.name,
        email: newUser.email,
        isAdmin,
        roles: Array.isArray(selectedRoleOption)
          ? selectedRoleOption.map((role: Options) => ({
              id: parseInt(role.value),
            }))
          : [],
      });

      const user = userResp.data;
      // Save user to state
      setUsers([...users, user]);

      // Send email for user to register a key
      await axios.post(`/users/register`, {
        email: newUser.email,
      });
      router.push("/admin");
    }
  };

  const newUserModalContent = (
    <div className={styles.modalContentContainer}>
      <h2 className={styles.subheader}>Name</h2>
      <input
        className={styles.modalInput}
        type="text"
        value={newUser.name}
        onChange={(e) => setNewUser({ ...newUser, name: e.target.value })}
      />
      <h2 className={styles.subheader}>Email</h2>
      <input
        className={styles.modalInput}
        type="text"
        value={newUser.email}
        onChange={(e) => setNewUser({ ...newUser, email: e.target.value })}
      />
      <SearchDropdown
        options={roles}
        placeholder="Select Role(s)..."
        subheader="Roles"
        selectDropdown="tableModal"
        setSelectedOption={setSelectedRoleOption}
        isMulti
      />
      <label className={styles.checkboxLabel}>
        <input
          className={styles.modalCheckbox}
          type="checkbox"
          checked={isAdmin}
          onChange={(e) => setIsAdmin(e.target.checked)}
        />
        <p>Admin</p>
      </label>
      <div className={styles.modalButtonGroup}>
        <SubmitButton
          disabled={disabled}
          text="Submit"
          onClick={handleModalSubmit}
        />
      </div>
    </div>
  );

  return (
    <Modal
      closeModal={handleModalClose}
      title="New User"
      content={newUserModalContent}
    />
  );
}
