"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { useAuth } from "@/lib/auth";
import { api } from "@/lib/api";
import { MemberForm } from "@/components/member-form";

export default function EditMemberPage() {
    const { id } = useParams();
    const { token } = useAuth();

    const [member, setMember] = useState<any>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadMember();
    }, []);

    async function loadMember() {
        try {
            const res = await api(`/members/${id}`, {
                token,
            });

            setMember(res.data);
        } catch (err) {
            console.error(err);
        } finally {
            setLoading(false);
        }
    }

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