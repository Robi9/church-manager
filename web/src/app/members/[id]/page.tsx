"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";
import {
    MemberForm,
    type MemberFormInitialData,
} from "@/components/member-form";

export default function EditMemberPage() {
    const { id } = useParams();

    const [member, setMember] = useState<MemberFormInitialData | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        let active = true;
        api<MemberFormInitialData>(`/members/${id}`)
            .then((res) => { if (active) setMember(res.data); })
            .catch(() => { if (active) setMember(null); })
            .finally(() => { if (active) setLoading(false); });
        return () => { active = false; };
    }, [id]);

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
