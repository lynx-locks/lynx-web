"use client";

import User from "@/types/user";
import styles from "./table.module.css";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faCaretDown,
  faCaretUp,
  faGear,
} from "@fortawesome/free-solid-svg-icons";
import { useEffect, useState } from "react";
import Modal from "@/components/modal/modal";
import { SubmitButton } from "@/components/button/button";
import SearchDropdown from "../searchDropdown/searchDropdown";
import { Options, SelectType } from "@/types/selectOptions";
import { getRoleOptions, getUserRoles } from "@/data/roles";
import Loader from "@/components/loader/loader";
import axios from "@/axios/client";

export default function AdminTable({
  users,
  updateUser,
  deleteUser,
}: {
  users: User[];
  updateUser: (user: User, roles: SelectType) => void;
  deleteUser: (user: User) => Promise<void>;
}) {
  const [settingsUser, setSettingsUser] = useState<User | null>(null);
  const [columnHeaders, setColumnHeaders] = useState([
    {
      name: "Name",
      sort: "",
    },
    {
      name: "Email",
      sort: "",
    },
    {
      name: "Last Time In",
      sort: "",
    },
    {
      name: "Date",
      sort: "",
    },
  ]);
  const [sortedUsers, setSortedUsers] = useState(users);
  const [selectedRoleOption, setSelectedRoleOption] =
    useState<SelectType>(null);
  const [defaultRoles, setDefaultRoles] = useState<Options[]>([]);
  const [roles, setRoles] = useState<Options[]>([]);

  useEffect(() => {
    const sortBy = columnHeaders.find((header) => header.sort !== "");
    const newUsers = [...users];
    newUsers.sort((a, b) => {
      const first = sortBy?.sort === "asc" ? a : b;
      const second = sortBy?.sort === "asc" ? b : a;
      if (sortBy?.name === "Last Time In") {
        return (
          (first.timeIn &&
            second.timeIn &&
            first.timeIn.localeCompare(second.timeIn)) ||
          (sortBy?.sort === "asc" ? 1 : -1)
        );
      } else if (sortBy?.name === "Date") {
        return (
          (first.lastDateIn &&
            second.lastDateIn &&
            first.lastDateIn.localeCompare(second.lastDateIn)) ||
          (sortBy?.sort === "asc" ? 1 : -1)
        );
      } else if (sortBy?.name === "Email") {
        return first.email.localeCompare(second.email);
      } else if (sortBy?.name === "Name") {
        return first.name.localeCompare(second.name);
      }
      return 0;
    });
    setSortedUsers(newUsers);
  }, [columnHeaders, setSortedUsers, users]);

  const handleClickSettings = async (idx: number) => {
    const user = sortedUsers[idx];
    setDefaultRoles(await getUserRoles(user.id));
    setRoles(await getRoleOptions());
    setSettingsUser(user);
  };

  const closeModal = () => {
    setSettingsUser(null);
  };

  const handleSort = (idx: number) => {
    const newColumnHeaders = columnHeaders.map((header, i) => {
      if (i === idx) {
        return {
          ...header,
          sort: header.sort === "asc" ? "desc" : "asc",
        };
      }
      return {
        ...header,
        sort: "",
      };
    });
    setColumnHeaders(newColumnHeaders);
  };

  const handleSettingsTextChange = (
    user: User,
    key: "name" | "email",
    value: string,
  ) => {
    setSettingsUser({ ...user, [key]: value });
  };

  const handleSubmitSettings = () => {
    updateUser(settingsUser!, selectedRoleOption);
    closeModal();
  };

  const handleDeleteUser = () => {
    if (confirm("Are you sure you want to delete this user?") && settingsUser) {
      deleteUser(settingsUser).then(() => closeModal());
    }
  };

  const handleRevokeKey = async () => {
    if (confirm("Are you sure you want to revoke this user's keys?")) {
      const resp = await axios.delete(`/users/${settingsUser?.id}/creds`);
      if (resp.status === 200) {
        alert("Keys revoked successfully");
      }
    }
  };

  return (
    <div className={styles.tableContainer}>
      {sortedUsers.length === 0 ? (
        <Loader />
      ) : (
        <table className={styles.table} border={1} rules="rows">
          <thead className={styles.tableHeader}>
            <tr>
              <th className={styles.tableCell}>
                <input type="checkbox" />
              </th>
              {columnHeaders.map((header, idx) => (
                <th key={idx} className={styles.tableCell}>
                  <p className={styles.columnHeader}>{header.name}</p>
                  <button onClick={() => handleSort(idx)}>
                    {header.sort === "asc" ? (
                      <FontAwesomeIcon icon={faCaretUp} size="lg" />
                    ) : (
                      <FontAwesomeIcon icon={faCaretDown} size="lg" />
                    )}
                  </button>
                </th>
              ))}
              <th />
            </tr>
          </thead>
          <tbody>
            {sortedUsers.map((user, idx) => (
              <tr
                key={user.id}
                className={
                  idx < users.length - 1 ? styles.tableRow : styles.tableLastRow
                }
              >
                <td className={styles.tableCell}>
                  <input type="checkbox" />
                </td>
                <td className={styles.tableCell}>{user.name}</td>
                <td className={styles.tableCell}>{user.email}</td>
                <td className={styles.tableCell}>{user.timeIn || "N/A"}</td>
                <td className={styles.tableCell}>{user.lastDateIn || "N/A"}</td>
                <td>
                  <FontAwesomeIcon
                    className={styles.settingsIcon}
                    icon={faGear}
                    onClick={() => handleClickSettings(idx)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {settingsUser && (
        <Modal
          closeModal={closeModal}
          title={`Settings for ${settingsUser.name}`}
          content={
            <div className={styles.settingsModal}>
              <div className={styles.settingsInputContainer}>
                <div className={styles.settingsInputLabel}>Name:</div>
                <input
                  className={styles.settingsInput}
                  type="text"
                  value={settingsUser.name}
                  onChange={(e) =>
                    handleSettingsTextChange(
                      settingsUser,
                      "name",
                      e.target.value,
                    )
                  }
                />
                <div className={styles.settingsInputLabel}>Email:</div>
                <input
                  className={styles.settingsInput}
                  type="text"
                  value={settingsUser.email}
                  onChange={(e) =>
                    handleSettingsTextChange(
                      settingsUser,
                      "email",
                      e.target.value,
                    )
                  }
                />
                <div className={styles.settingsInputLabel}>Roles:</div>
                <SearchDropdown
                  defaultValue={defaultRoles}
                  options={roles}
                  placeholder="Select Role(s)..."
                  subheader=""
                  setSelectedOption={setSelectedRoleOption}
                  selectDropdown="settingsModal"
                  isMulti
                />
              </div>
              <div className={styles.settingsButtonGroup}>
                <button
                  className={styles.deleteButton}
                  onClick={handleRevokeKey}
                >
                  Revoke Key
                </button>
                <button
                  className={styles.deleteButton}
                  onClick={handleDeleteUser}
                >
                  Delete User
                </button>
              </div>
              <SubmitButton
                text="Submit Changes"
                onClick={handleSubmitSettings}
              />
            </div>
          }
        />
      )}
    </div>
  );
}
