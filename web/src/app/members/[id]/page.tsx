"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { api } from "@/lib/api";
import {
    MemberForm,
    type MemberFormInitialData,
} from "@/components/member-form";

export default function EditMemberPage() {
    const { id } = useParams();
    const { token } = useAuth();

    const [member, setMember] = useState<MemberFormInitialData | null>(null);
    const [loading, setLoading] = useState(true);

    const loadMember = useCallback(async () => {
        try {
            const res = await api<MemberFormInitialData>(`/members/${id}`, {
                token,
            });

            setMember(res.data);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    }, [id, token]);

    useEffect(() => {
        void loadMember();
    }, [loadMember]);

    if (loading) {
        return <div>Carregando...</div>;
    }

    if (!member) {
        return <div>Membro não encontrado</div>;
    }

    return (
        <MemberForm
            mode="edit"
            initialData={member}
        />
    );
}
